package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dapr/go-sdk/actor"
	actorErr "github.com/dapr/go-sdk/actor/error"
	"github.com/dapr/go-sdk/actor/reminder"
	"github.com/dapr/go-sdk/actor/timer"
	perrors "github.com/pkg/errors"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

// ActorManager is to manage one type of actor
type ActorManager struct {
	// factory stores the actor factory of specific type of actor
	factory actor.ActorImplFactory

	service actor.ActorImpl

	// activeActors stores the map actorID -> ActorContainer
	activeActors sync.Map
}

func NewActorManager() *ActorManager {
	return &ActorManager{}
}

// RegisterActorImplFactory registers the action factory @f
func (m *ActorManager) RegisterActorImplFactory(f actor.ActorImplFactory) {
	m.factory = f
}

// InvokeMethod to invoke local function by @actorID, @methodName and @request request param
func (m *ActorManager) InvokeMethod(actorID, methodName string, request []byte) ([]byte, actorErr.ActorError) {
	val, ok := m.activeActors.Load(actorID)
	if !ok {
		m.service = m.factory()
		m.activeActors.Store(actorID, NewActorContainer(m.service))
		val, _ = m.activeActors.Load(actorID)
	}
	methodType := val.(*ActorContainer).methodType[methodName]
	argsValues := make([]reflect.Value, 0)
	argsValues = append(argsValues, reflect.ValueOf(m.service))
	argsValues = append(argsValues, reflect.ValueOf(context.Background()))
	var replyv reflect.Value
	if len(methodType.ArgsType()) > 0 {
		typ := methodType.ArgsType()[0]
		paramValue := reflect.New(typ)
		paramInterface := paramValue.Interface()
		if err := json.Unmarshal(request, paramInterface); err != nil {
			return nil, actorErr.ErrorActorInvokeFailed
		}
		argsValues = append(argsValues, reflect.ValueOf(paramInterface).Elem())
	}
	returnValue := methodType.Method().Func.Call(argsValues)
	var retErr interface{}
	if len(returnValue) == 1 {
		return nil, actorErr.Success
	}

	if len(returnValue) == 2 {
		replyv = returnValue[0]
		retErr = returnValue[1].Interface()
	}

	if retErr != nil {
		panic(retErr)
	}
	rspData, err := json.Marshal(replyv.Interface())
	if err != nil {
		return nil, actorErr.ErrorActorInvokeFailed
	}
	return rspData, actorErr.Success
}

func (m *ActorManager) DetectiveActor(actorID string) actorErr.ActorError {
	val, ok := m.activeActors.Load(actorID)
	if !ok {
		return actorErr.ErrorActorIDNotFound
	}
	val.(actor.ActorImpl).OnDeactive()
	return actorErr.Success
}

func (m *ActorManager) ActiveManager(actorID string) {
	_, ok := m.activeActors.Load(actorID)
	if ok {
		return
	}
	// todo create actor
	//m.activeActors.Store(actor.)
}

func (m *ActorManager) InvokeReminder(actorID, reminderName string, params []byte) actorErr.ActorError {
	val, ok := m.activeActors.Load(actorID)
	if !ok {
		return actorErr.ErrorActorIDNotFound
	}
	reminderParams := &reminder.ActorReminderParams{}
	if err := json.Unmarshal(params, reminderParams); err != nil {
		fmt.Println("unmarshal reminder param error = ", err)
		return actorErr.ErrorRemindersParamsInvalid
	}

	val.(actor.ActorImpl).ReceiveReminder(reminderName, reminderParams.Data, reminderParams.DueTime, reminderParams.Period)
	return actorErr.Success
}

func (m *ActorManager) InvokeTimer(actorID, timerName string, params []byte) actorErr.ActorError {
	//val, ok := m.activeActors.Load(actorID)
	//if !ok {
	//	return actorErr.ErrorActorIDNotFound
	//}
	timerParams := &timer.ActorTimerParam{}
	if err := json.Unmarshal(params, timerParams); err != nil {
		fmt.Println("unmarshal reminder param error = ", err)
		return actorErr.ErrorRemindersParamsInvalid
	}

	fmt.Println("timer param: call back = ", timerParams.CallBack, " data = ", string(timerParams.Data))

	// todo call back to target function
	//val.(actor.ActorImpl).Invoke(params)
	return actorErr.Success
}

func getAbsctractMethodMap(rcvr interface{}) map[string]*MethodType {
	s := new(Service)
	s.rcvrType = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if sname == "" {
		panic("sname == empty")
	}
	if !isExported(sname) {
		s := "type " + sname + " is not exported"
		panic(s)
	}

	return suitableMethods(s.rcvrType)
}

// Service is description of service
type Service struct {
	name     string
	rcvr     reflect.Value
	rcvrType reflect.Type
	methods  map[string]*MethodType
}

// Method gets @s.methods.
func (s *Service) Method() map[string]*MethodType {
	return s.methods
}

// Name will return service name
func (s *Service) Name() string {
	return s.name
}

// RcvrType gets @s.rcvrType.
func (s *Service) RcvrType() reflect.Type {
	return s.rcvrType
}

// Rcvr gets @s.rcvr.
func (s *Service) Rcvr() reflect.Value {
	return s.rcvr
}

// Is this an exported - upper case - name
func isExported(name string) bool {
	s, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(s)
}

// MethodType is description of service method.
type MethodType struct {
	method    reflect.Method
	ctxType   reflect.Type   // request context
	argsType  []reflect.Type // args except ctx, include replyType if existing
	replyType reflect.Type   // return value, otherwise it is nil
}

// Method gets @m.method.
func (m *MethodType) Method() reflect.Method {
	return m.method
}

// CtxType gets @m.ctxType.
func (m *MethodType) CtxType() reflect.Type {
	return m.ctxType
}

// ArgsType gets @m.argsType.
func (m *MethodType) ArgsType() []reflect.Type {
	return m.argsType
}

// ReplyType gets @m.replyType.
func (m *MethodType) ReplyType() reflect.Type {
	return m.replyType
}

// suitableMethods returns suitable Rpc methods of typ
func suitableMethods(typ reflect.Type) map[string]*MethodType {
	methods := make(map[string]*MethodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		if mt, err := suiteMethod(method); mt != nil && err == nil {
			methods[method.Name] = mt
		}
	}
	return methods
}

// suiteMethod returns a suitable Rpc methodType
func suiteMethod(method reflect.Method) (*MethodType, error) {
	mtype := method.Type
	mname := method.Name
	inNum := mtype.NumIn()
	outNum := mtype.NumOut()

	// Method must be exported.
	if method.PkgPath != "" {
		return nil, perrors.New("method is not exported")
	}

	var (
		replyType, ctxType reflect.Type
		argsType           []reflect.Type
	)

	if outNum != 1 && outNum != 2 {
		return nil, perrors.New("num out invalid")
	}

	// The latest return type of the method must be error.
	if returnType := mtype.Out(outNum - 1); returnType != typeOfError {
		return nil, perrors.New(fmt.Sprintf("the latest return type %s of method %q is not error", returnType, mname))
	}

	// replyType
	if outNum == 2 {
		replyType = mtype.Out(0)
		if !isExportedOrBuiltinType(replyType) {
			return nil, perrors.New(fmt.Sprintf("reply type of method %s not exported{%v}", mname, replyType))
		}
	}

	index := 1

	// ctxType
	if inNum > 1 && mtype.In(1).String() == "context.Context" {
		ctxType = mtype.In(1)
		index = 2
	}

	for ; index < inNum; index++ {
		argsType = append(argsType, mtype.In(index))
		// need not be a pointer.
		if !isExportedOrBuiltinType(mtype.In(index)) {
			return nil, perrors.New(fmt.Sprintf("argument type of method %q is not exported %v", mname, mtype.In(index)))
		}
	}

	return &MethodType{method: method, argsType: argsType, replyType: replyType, ctxType: ctxType}, nil
}

var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}
