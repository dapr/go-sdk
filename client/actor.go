package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dapr/go-sdk/actor"
	pb "github.com/dapr/go-sdk/dapr/proto/runtime/v1"
	"github.com/pkg/errors"
	"reflect"
)

type InvokeActorRequest struct {
	ActorType string
	ActorID   string
	Method    string
	Data      []byte
}

type InvokeActorResponse struct {
	Data []byte
}

func (c *GRPCClient) ImplActorInteface(actorInterface actor.ActorProxy) {
	ImplActor(actorInterface, c)
}

// InvokeActor invokes specific operation on the configured Dapr binding.
// This method covers input, output, and bi-directional bindings.
func (c *GRPCClient) InvokeActor(ctx context.Context, in *InvokeActorRequest) (out *InvokeActorResponse, err error) {
	// todo param check
	//if in == nil {
	//	return nil, errors.New("binding invocation required")
	//}
	//if in.Method == "" {
	//	return nil, errors.New("binding invocation name required")
	//}
	//if in.ActorType == "" {
	//	return nil, errors.New("binding invocation operation required")
	//}

	req := &pb.InvokeActorRequest{
		ActorType: in.ActorType,
		ActorId:   in.ActorID,
		Method:    in.Method,
		Data:      in.Data,
	}

	resp, err := c.protoClient.InvokeActor(c.withAuthToken(ctx), req)
	if err != nil {
		return nil, errors.Wrapf(err, "error invoking binding %s/%s", in.ActorType, in.ActorID)
	}

	out = &InvokeActorResponse{}

	if resp != nil {
		out.Data = resp.Data
	}

	return out, nil
}

type RegisterActorReminderRequest struct {
	ActorType string
	ActorID   string
	Name      string
	DueTime   string
	Period    string
	Data      []byte
}

// RegisterActorReminder invokes specific operation on the configured Dapr binding.
// This method covers input, output, and bi-directional bindings.
func (c *GRPCClient) RegisterActorReminder(ctx context.Context, in *RegisterActorReminderRequest) (err error) {
	// todo param check
	if in == nil {
		return errors.New("binding invocation required")
	}
	//if in.Method == "" {
	//	return  errors.New("binding invocation name required")
	//}
	if in.ActorType == "" {
		return errors.New("binding invocation operation required")
	}

	req := &pb.RegisterActorReminderRequest{
		ActorType: in.ActorType,
		ActorId:   in.ActorID,
		Name:      in.Name,
		DueTime:   in.DueTime,
		Period:    in.Period,
		Data:      in.Data,
	}

	_, err = c.protoClient.RegisterActorReminder(c.withAuthToken(ctx), req)
	if err != nil {
		return errors.Wrapf(err, "error invoking binding %s/%s", in.ActorType, in.ActorID)
	}

	return
}

type RegisterActorTimerRequest struct {
	ActorType string
	ActorID   string
	Name      string
	DueTime   string
	Period    string
	Data      []byte
	CallBack  string
}

// RegisterActorTimer invokes specific operation on the configured Dapr binding.
// This method covers input, output, and bi-directional bindings.
func (c *GRPCClient) RegisterActorTimer(ctx context.Context, in *RegisterActorTimerRequest) (err error) {
	// todo param check
	if in == nil {
		return errors.New("binding invocation required")
	}
	//if in.Method == "" {
	//	return  errors.New("binding invocation name required")
	//}
	if in.ActorType == "" {
		return errors.New("binding invocation operation required")
	}

	req := &pb.RegisterActorTimerRequest{
		ActorType: in.ActorType,
		ActorId:   in.ActorID,
		Name:      in.Name,
		DueTime:   in.DueTime,
		Period:    in.Period,
		Data:      in.Data,
		Callback:  in.CallBack,
	}

	_, err = c.protoClient.RegisterActorTimer(c.withAuthToken(ctx), req)
	if err != nil {
		return errors.Wrapf(err, "error invoking binding %s/%s", in.ActorType, in.ActorID)
	}

	return
}

func (c *GRPCClient) SaveStateTransactionally(ctx context.Context, actorID, actorType string, ops []*pb.TransactionalActorStateOperation) error {
	req := &pb.ExecuteActorStateTransactionRequest{
		ActorId:    actorID,
		ActorType:  actorType,
		Operations: ops,
	}
	_, err := c.protoClient.ExecuteActorStateTransaction(c.withAuthToken(ctx), req)
	return err
}

func ImplActor(actor actor.ActorProxy, c Client) {
	actorValue := reflect.ValueOf(actor)
	fmt.Println("[Implement] reflect.TypeOf: ", actorValue.String())

	valueOfActor := actorValue.Elem()
	typeOfActor := valueOfActor.Type()

	// check incoming interface, incoming interface's elem must be a struct.
	if typeOfActor.Kind() != reflect.Struct {
		fmt.Println("%s must be a struct ptr", valueOfActor.String())
		return
	}

	numField := valueOfActor.NumField()
	for i := 0; i < numField; i++ {
		t := typeOfActor.Field(i)
		methodName := t.Name
		if methodName == "Type" {
			continue
		}
		f := valueOfActor.Field(i)
		if f.Kind() == reflect.Func && f.IsValid() && f.CanSet() {
			outNum := t.Type.NumOut()

			if outNum != 1 && outNum != 2 {
				fmt.Printf("method %s of mtype %v has wrong number of in out parameters %d; needs exactly 1/2\n",
					t.Name, t.Type.String(), outNum)
				continue
			}

			// The latest return type of the method must be error.
			if returnType := t.Type.Out(outNum - 1); returnType != reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()).Type() {
				fmt.Printf("the latest return type %s of method %q is not error\n", returnType, t.Name)
				continue
			}

			funcOuts := make([]reflect.Type, outNum)
			for i := 0; i < outNum; i++ {
				funcOuts[i] = t.Type.Out(i)
			}

			// do method proxy here:
			f.Set(reflect.MakeFunc(f.Type(), MakeCallProxyFunction(actor, methodName, funcOuts, c)))
			fmt.Printf("set method [%s]\n", methodName)
		}
	}

}

func MakeCallProxyFunction(actor actor.ActorProxy, methodName string, outs []reflect.Type, c Client) func(in []reflect.Value) []reflect.Value {
	return func(in []reflect.Value) []reflect.Value {
		var (
			err    error
			inIArr []interface{}
			reply  reflect.Value
		)

		if len(outs) == 2 {
			if outs[0].Kind() == reflect.Ptr {
				reply = reflect.New(outs[0].Elem())
			} else {
				reply = reflect.New(outs[0])
			}
		}

		start := 0
		end := len(in)
		invCtx := context.Background()
		if end > 0 {
			if in[0].Type().String() == "context.Context" {
				if !in[0].IsNil() {
					// the user declared context as method's parameter
					invCtx = in[0].Interface().(context.Context)
				}
				start += 1
			}
			if len(outs) == 1 && in[end-1].Type().Kind() == reflect.Ptr {
				end -= 1
				reply = in[len(in)-1]
			}
		}

		if end-start <= 0 {
			inIArr = []interface{}{}
		} else if end-start == 1 {
			inIArr = []interface{}{in[start].Interface()}
		} else {
			panic("param nums is zero or one is allowed by actor")
		}

		var data []byte
		if len(inIArr) > 0 {
			data, err = json.Marshal(inIArr[0])
		}
		if err != nil {
			panic(err)
		}

		rsp, err := c.InvokeActor(invCtx, &InvokeActorRequest{
			ActorType: actor.Type(),
			ActorID:   "testActorID", // todo geneate actor id
			Method:    methodName,
			Data:      data, // todo serialize
		})

		if len(outs) == 1 {
			return []reflect.Value{reflect.ValueOf(&err).Elem()}
		}

		response := reply.Interface()
		if rsp != nil {
			fmt.Println("invoke actor response = ", string(rsp.Data))
			if err := json.Unmarshal(rsp.Data, response); err != nil {
				fmt.Println(err)
			}
		}
		if len(outs) == 2 && outs[0].Kind() != reflect.Ptr {
			return []reflect.Value{reply.Elem(), reflect.ValueOf(&err).Elem()}
		}
		return []reflect.Value{reply, reflect.ValueOf(&err).Elem()}
	}
}
