/*
Copyright 2026 The Dapr Authors
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This example demonstrates workflow history propagation in a credit card
// payment processing scenario. A merchant payment workflow propagates its
// execution history to downstream fraud detection and payment gateway
// services, which inspect the history to make authorization decisions
// and enforce compliance guardrails.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dapr/durabletask-go/api/protos"
	"github.com/dapr/durabletask-go/workflow"
	"github.com/dapr/go-sdk/client"
)

// PaymentRequest represents a credit card payment to process.
type PaymentRequest struct {
	CardLast4   string  `json:"cardLast4"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	MerchantID  string  `json:"merchantId"`
	Description string  `json:"description"`
}

// FraudCheckResult is the output of the fraud detection workflow.
type FraudCheckResult struct {
	RiskScore  float64 `json:"riskScore"`
	Approved   bool    `json:"approved"`
	Reason     string  `json:"reason"`
	EventCount int     `json:"eventCount"`
}

// SettlementResult is the output of the settlement activity.
type SettlementResult struct {
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	EventCount    int    `json:"eventCount"`
}

var logger = log.New(os.Stdout, "", log.LstdFlags)

func main() {
	r := workflow.NewRegistry()

	// Step 1: MerchantCheckout (root workflow)
	//   Step 1.1: ValidateMerchant (activity)
	//   Step 1.2: ProcessPayment (child wf, propagate lineage)
	//     Step 1.2.1: ValidateCard (activity)
	//     Step 1.2.2: CheckSpendingLimits (activity)
	//     Step 1.2.3: FraudDetection (grandchild wf, propagate lineage)
	//     Step 1.2.4: SettlePayment (ProcessPayment activity, propagate own history)
	if err := r.AddWorkflow(MerchantCheckout); err != nil {
		logger.Fatal(err)
	}
	if err := r.AddActivity(ValidateMerchant); err != nil {
		logger.Fatal(err)
	}
	if err := r.AddWorkflow(ProcessPayment); err != nil {
		logger.Fatal(err)
	}
	if err := r.AddActivity(ValidateCard); err != nil {
		logger.Fatal(err)
	}
	if err := r.AddActivity(CheckSpendingLimits); err != nil {
		logger.Fatal(err)
	}
	if err := r.AddWorkflow(FraudDetection); err != nil {
		logger.Fatal(err)
	}
	if err := r.AddActivity(SettlePayment); err != nil {
		logger.Fatal(err)
	}

	wclient, err := client.NewWorkflowClient()
	if err != nil {
		logger.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err = wclient.StartWorker(ctx, r); err != nil {
		logger.Fatal(err)
	}

	fmt.Println(banner("WORKFLOW HISTORY PROPAGATION DEMO"))
	fmt.Println()
	fmt.Println("  Flow: MerchantCheckout -> ValidateMerchant")
	fmt.Println("           -> ProcessPayment (child wf, lineage)")
	fmt.Println("               -> ValidateCard -> CheckSpendingLimits")
	fmt.Println("               -> FraudDetection (child wf, lineage)    <-- sees MerchantCheckout + ProcessPayment events")
	fmt.Println("               -> SettlePayment (activity, own history) <-- sees only ProcessPayment events")
	fmt.Println()

	id, err := wclient.ScheduleWorkflow(ctx, "MerchantCheckout",
		workflow.WithInstanceID("checkout-001"),
		workflow.WithInput(PaymentRequest{
			CardLast4:   "4242",
			Amount:      149.99,
			Currency:    "USD",
			MerchantID:  "merchant-abc",
			Description: "Online purchase",
		}),
	)
	if err != nil {
		logger.Fatalf("failed to start workflow: %v", err)
	}
	fmt.Printf("  [main] Started workflow: %s\n", id)

	waitCtx, waitCancel := context.WithTimeout(ctx, 30*time.Second)
	_, err = wclient.WaitForWorkflowCompletion(waitCtx, id)
	waitCancel()
	if err != nil {
		logger.Fatalf("workflow failed: %v", err)
	}

	if err = wclient.PurgeWorkflowState(ctx, id); err != nil {
		logger.Printf("failed to purge: %v", err)
	}

	fmt.Println()
	fmt.Println(banner("COMPLETE"))

	// Block until SIGINT/SIGTERM so the K8s pod stays Ready and `kubectl logs`
	// keeps showing the demo. In standalone, Ctrl+C cancels ctx, the workflow
	// worker stops cleanly, and this returns.
	<-ctx.Done()
}

// MerchantCheckout is the top-level workflow. It validates the merchant,
// then calls ProcessPayment as a child workflow with PropagateLineage(),
// giving ProcessPayment ancestral history to forward (or not) downstream.
func MerchantCheckout(ctx *workflow.WorkflowContext) (any, error) {
	var req PaymentRequest
	if err := ctx.GetInput(&req); err != nil {
		return nil, err
	}

	if !ctx.IsReplaying() {
		fmt.Printf("  [MerchantCheckout] Starting checkout for merchant %s\n", req.MerchantID)
		fmt.Println("  [MerchantCheckout] Step 1: CallActivity(ValidateMerchant) — no propagation")
	}
	var merchantValid bool
	if err := ctx.CallActivity(ValidateMerchant,
		workflow.WithActivityInput(req),
	).Await(&merchantValid); err != nil {
		return nil, fmt.Errorf("merchant validation failed: %w", err)
	}
	if !ctx.IsReplaying() {
		fmt.Println("  [MerchantCheckout] Step 1 complete: merchant valid")
		fmt.Println("  [MerchantCheckout] Step 2: CallChildWorkflow(ProcessPayment)")
		fmt.Println("                     -> WithHistoryPropagation(PropagateLineage)")
	}
	var result string
	if err := ctx.CallChildWorkflow(ProcessPayment,
		workflow.WithChildWorkflowInput(req),
		workflow.WithHistoryPropagation(
			workflow.PropagateLineage()),
	).Await(&result); err != nil {
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}
	if !ctx.IsReplaying() {
		fmt.Printf("  [MerchantCheckout] COMPLETE: %s\n", result)
	}
	return result, nil
}

// ValidateMerchant checks if the merchant is registered and in good standing.
func ValidateMerchant(ctx workflow.ActivityContext) (any, error) {
	var req PaymentRequest
	if err := ctx.GetInput(&req); err != nil {
		return false, err
	}
	fmt.Printf("  [ValidateMerchant] Validating merchant %s\n", req.MerchantID)
	return true, nil
}

// ProcessPayment orchestrates a credit card payment. It validates the card,
// runs fraud detection (as a child wf with full history propagation),
// and settles the payment (as an activity with workflow-level propagation).
func ProcessPayment(ctx *workflow.WorkflowContext) (any, error) {
	var req PaymentRequest
	if err := ctx.GetInput(&req); err != nil {
		return nil, err
	}

	if !ctx.IsReplaying() {
		fmt.Printf("  [ProcessPayment] Starting payment: ****%s, %.2f %s\n",
			req.CardLast4, req.Amount, req.Currency)
	}

	// Step 1: Validate the credit card (no propagation — plain activity)
	if !ctx.IsReplaying() {
		fmt.Println("  [ProcessPayment] Step 1: CallActivity(ValidateCard) — no propagation")
	}
	var cardValid bool
	if err := ctx.CallActivity(ValidateCard,
		workflow.WithActivityInput(req),
	).Await(&cardValid); err != nil {
		return nil, fmt.Errorf("card validation failed: %w", err)
	}
	if !cardValid {
		return "payment declined: invalid card", nil
	}
	if !ctx.IsReplaying() {
		fmt.Println("  [ProcessPayment] Step 1 complete: card valid")
	}

	// Step 2: Check spending limits (no propagation — plain activity)
	if !ctx.IsReplaying() {
		fmt.Println("  [ProcessPayment] Step 2: CallActivity(CheckSpendingLimits) — no propagation")
	}
	var withinLimits bool
	if err := ctx.CallActivity(CheckSpendingLimits,
		workflow.WithActivityInput(req),
	).Await(&withinLimits); err != nil {
		return nil, fmt.Errorf("spending limit check failed: %w", err)
	}
	if !withinLimits {
		return "payment declined: spending limit exceeded", nil
	}
	if !ctx.IsReplaying() {
		fmt.Println("  [ProcessPayment] Step 2 complete: within limits")
	}

	// Step 3: Run fraud detection as a child wf.
	// PropagateLineage() — include our events AND any ancestral history we received
	if !ctx.IsReplaying() {
		fmt.Println("  [ProcessPayment] Step 3: CallChildWorkflow(FraudDetection)")
		fmt.Println("                   -> WithHistoryPropagation(PropagateLineage)")
	}
	var fraudResult FraudCheckResult
	if err := ctx.CallChildWorkflow(FraudDetection,
		workflow.WithChildWorkflowInput(req),
		workflow.WithHistoryPropagation(workflow.PropagateLineage()), // 15 events, 0-14 (includes ValidateMerchant from parent + ValidateCard + CheckSpendingLimits from self which is the ProcessPayment wf)
	).Await(&fraudResult); err != nil {
		return nil, fmt.Errorf("fraud detection failed: %w", err)
	}
	if !fraudResult.Approved {
		return fmt.Sprintf("payment declined: fraud check failed (risk=%.2f, reason=%s)",
			fraudResult.RiskScore, fraudResult.Reason), nil
	}
	if !ctx.IsReplaying() {
		fmt.Printf("  [ProcessPayment] Step 3 complete: fraud check passed (risk=%.2f, %d events verified)\n",
			fraudResult.RiskScore, fraudResult.EventCount)
	}

	// Step 4: Settle the payment.
	// PropagateOwnHistory() — include our events only (no ancestral chain)
	if !ctx.IsReplaying() {
		fmt.Println("  [ProcessPayment] Step 4: CallActivity(SettlePayment)")
		fmt.Println("                   -> WithHistoryPropagation(PropagateOwnHistory)")
	}
	var settlement SettlementResult
	if err := ctx.CallActivity(SettlePayment,
		workflow.WithActivityInput(req),
		workflow.WithHistoryPropagation(workflow.PropagateOwnHistory()), // 12 events, 0-11 (only processPayment wf, does not include ancestral lineage from MerchantCheckout wf)
	).Await(&settlement); err != nil {
		return nil, fmt.Errorf("settlement failed: %w", err)
	}
	if !ctx.IsReplaying() {
		fmt.Printf("  [ProcessPayment] Step 4 complete: settled (txn=%s, %d events verified)\n",
			settlement.TransactionID, settlement.EventCount)
	}

	result := fmt.Sprintf("payment settled: txn=%s, card=****%s, amount=%.2f %s",
		settlement.TransactionID, req.CardLast4, req.Amount, req.Currency)
	if !ctx.IsReplaying() {
		fmt.Printf("  [ProcessPayment] COMPLETE: %s\n", result)
	}
	return result, nil
}

// FraudDetection is a child wf that inspects the parent's propagated
// history to make risk decisions. It verifies that the required upstream
// steps (card validation, spending limits) were executed before scoring.
func FraudDetection(ctx *workflow.WorkflowContext) (any, error) {
	var req PaymentRequest
	if err := ctx.GetInput(&req); err != nil {
		return nil, err
	}

	fmt.Printf("  [FraudDetection] Checking payment: ****%s, %.2f %s\n",
		req.CardLast4, req.Amount, req.Currency)

	history := ctx.GetPropagatedHistory()
	if history == nil {
		fmt.Println("  [FraudDetection] WARNING: No propagated history received!")
		fmt.Println("  [FraudDetection] DENIED — cannot verify caller pipeline without history")
		return FraudCheckResult{
			RiskScore: 1.0,
			Approved:  false,
			Reason:    "no execution history provided — cannot verify caller pipeline",
		}, nil
	}

	fmt.Printf("  [FraudDetection] Received propagated history: %d events (scope: %s)\n",
		len(history.Events()), describeScope(history.Scope()))
	fmt.Printf("  [FraudDetection] Apps in chain: %v\n", history.GetAppIDs())
	for _, wf := range history.GetWorkflows() {
		fmt.Printf("  [FraudDetection]   workflow: app=%s, name=%s, instance=%s\n",
			wf.AppID, wf.Name, wf.InstanceID)
	}

	merchantWf, err := history.GetWorkflowByName("MerchantCheckout")
	if err != nil {
		return FraudCheckResult{}, fmt.Errorf("expected MerchantCheckout in propagated history: %w", err)
	}
	processPaymentWf, err := history.GetWorkflowByName("ProcessPayment")
	if err != nil {
		return FraudCheckResult{}, fmt.Errorf("expected ProcessPayment in propagated history: %w", err)
	}
	merchant, err := merchantWf.GetActivityByName("ValidateMerchant")
	if err != nil {
		return FraudCheckResult{}, fmt.Errorf("expected ValidateMerchant in propagated history: %w", err)
	}
	card, err := processPaymentWf.GetActivityByName("ValidateCard")
	if err != nil {
		return FraudCheckResult{}, fmt.Errorf("expected ValidateCard in propagated history: %w", err)
	}
	spending, err := processPaymentWf.GetActivityByName("CheckSpendingLimits")
	if err != nil {
		return FraudCheckResult{}, fmt.Errorf("expected CheckSpendingLimits in propagated history: %w", err)
	}

	fmt.Printf("  [FraudDetection] Verification:\n")
	fmt.Printf("  [FraudDetection]   MerchantCheckout/ValidateMerchant: started=%v, completed=%v\n",
		merchant.Started, merchant.Completed)
	fmt.Printf("  [FraudDetection]   ProcessPayment/ValidateCard: started=%v, completed=%v\n",
		card.Started, card.Completed)
	fmt.Printf("  [FraudDetection]   ProcessPayment/CheckSpendingLimits: started=%v, completed=%v\n",
		spending.Started, spending.Completed)

	if !merchant.Completed || !card.Completed || !spending.Completed {
		fmt.Println("  [FraudDetection] DENIED — required upstream checks not completed")
		return FraudCheckResult{
			RiskScore:  0.9,
			Approved:   false,
			Reason:     "required upstream checks not completed in propagated history",
			EventCount: len(history.Events()),
		}, nil
	}

	riskScore := 0.1
	if req.Amount > 1000 {
		riskScore = 0.3
	}

	fmt.Printf("  [FraudDetection] APPROVED (risk=%.2f)\n", riskScore)
	return FraudCheckResult{
		RiskScore:  riskScore,
		Approved:   true,
		Reason:     "all upstream checks verified in propagated history",
		EventCount: len(history.Events()),
	}, nil
}

// ValidateCard checks if the credit card is valid.
func ValidateCard(ctx workflow.ActivityContext) (any, error) {
	var req PaymentRequest
	if err := ctx.GetInput(&req); err != nil {
		return false, err
	}

	ph := ctx.GetPropagatedHistory()
	fmt.Printf("  [ValidateCard] Validating card ****%s (propagated history: %s)\n",
		req.CardLast4, describeHistory(ph))
	return true, nil
}

// CheckSpendingLimits verifies the transaction is within the card's limits.
func CheckSpendingLimits(ctx workflow.ActivityContext) (any, error) {
	var req PaymentRequest
	if err := ctx.GetInput(&req); err != nil {
		return false, err
	}

	ph := ctx.GetPropagatedHistory()
	fmt.Printf("  [CheckSpendingLimits] Checking %.2f %s (propagated history: %s)\n",
		req.Amount, req.Currency, describeHistory(ph))
	return req.Amount <= 10000, nil
}

// SettlePayment processes the final payment settlement. It inspects the
// propagated workflow history to verify the payment pipeline executed
// correctly before settling.
func SettlePayment(ctx workflow.ActivityContext) (any, error) {
	var req PaymentRequest
	if err := ctx.GetInput(&req); err != nil {
		return nil, err
	}

	ph := ctx.GetPropagatedHistory()
	fmt.Printf("  [SettlePayment] Settling %.2f %s for merchant %s (propagated history: %s)\n",
		req.Amount, req.Currency, req.MerchantID, describeHistory(ph))

	eventCount := 0
	if ph != nil {
		eventCount = len(ph.Events())
		fmt.Printf("  [SettlePayment] Apps in chain: %v\n", ph.GetAppIDs())
		for _, wf := range ph.GetWorkflows() {
			fmt.Printf("  [SettlePayment]   workflow: app=%s, name=%s, instance=%s\n",
				wf.AppID, wf.Name, wf.InstanceID)
		}
		scheduledNames := make(map[string]string) // key=taskExecutionId, val=activity name
		for i, event := range ph.Events() {
			if ts := event.GetTaskScheduled(); ts != nil {
				scheduledNames[ts.GetTaskExecutionId()] = ts.GetName()
			}
			fmt.Printf("  [SettlePayment]   event[%d]: %s\n", i, describeEventResolved(event, scheduledNames))
		}
	}

	txnID := fmt.Sprintf("txn-%s-%d", req.MerchantID, time.Now().UnixMilli())
	fmt.Printf("  [SettlePayment] SETTLED: %s\n", txnID)

	return SettlementResult{
		TransactionID: txnID,
		Status:        "settled",
		EventCount:    eventCount,
	}, nil
}

// describeEventResolved returns a human-readable description of a history event
func describeEventResolved(event *protos.HistoryEvent, scheduledNames map[string]string) string {
	eventType := fmt.Sprintf("%T", event.EventType)
	if idx := strings.LastIndex(eventType, "."); idx >= 0 {
		eventType = eventType[idx+1:]
	}
	switch {
	case event.GetTaskScheduled() != nil:
		return fmt.Sprintf("%s -> %s", eventType, event.GetTaskScheduled().GetName())
	case event.GetTaskCompleted() != nil:
		if name, ok := scheduledNames[event.GetTaskCompleted().GetTaskExecutionId()]; ok {
			return fmt.Sprintf("%s -> %s", eventType, name)
		}
		return eventType
	case event.GetExecutionStarted() != nil:
		return fmt.Sprintf("%s -> %s", eventType, event.GetExecutionStarted().GetName())
	case event.GetChildWorkflowInstanceCreated() != nil:
		return fmt.Sprintf("%s -> %s", eventType, event.GetChildWorkflowInstanceCreated().GetName())
	default:
		return eventType
	}
}

func describeHistory(ph *workflow.PropagatedHistory) string {
	if ph == nil {
		return "none"
	}
	return fmt.Sprintf("%d events, scope=%s", len(ph.Events()), describeScope(ph.Scope()))
}

func describeScope(scope fmt.Stringer) string {
	s := scope.String()
	s = strings.TrimPrefix(s, "HISTORY_PROPAGATION_SCOPE_")
	return s
}

func banner(msg string) string {
	line := strings.Repeat("=", len(msg)+4)
	return fmt.Sprintf("%s\n= %s =\n%s", line, msg, line)
}
