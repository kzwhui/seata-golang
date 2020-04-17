package tm

import (
	"context"
	"reflect"
)

import (
	"github.com/pkg/errors"
)

import (
	context2 "github.com/dk-lockdown/seata-golang/client/context"
	"github.com/dk-lockdown/seata-golang/client/proxy"
	"github.com/dk-lockdown/seata-golang/pkg/logging"
)

type GlobalTransactionProxyService interface {
	GetProxyService() proxy.ProxyService
	GetMethodTransactionInfo(methodName string) *TransactionInfo
}

var (
	typError = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem()).Type()
)

func Implement(v GlobalTransactionProxyService) {
	// check parameters, incoming interface must be a elem's pointer.
	valueOf := reflect.ValueOf(v)
	logging.Logger.Debugf("[Implement] reflect.TypeOf: %s", valueOf.String())

	valueOfElem := valueOf.Elem()
	typeOf := valueOfElem.Type()

	// check incoming interface, incoming interface's elem must be a struct.
	if typeOf.Kind() != reflect.Struct {
		logging.Logger.Errorf("%s must be a struct ptr", valueOf.String())
		return
	}
	proxiedService := v.GetProxyService()
	pxdService := reflect.ValueOf(proxiedService)
	serviceName := reflect.Indirect(pxdService).Type().Name()

	makeCallProxy := func(serviceName, methodName string,txInfo *TransactionInfo) func(in []reflect.Value) []reflect.Value {
		return func(in []reflect.Value) []reflect.Value {
			var (
				args         = make([]interface{},0)
				returnValues = make([]reflect.Value,0)
				suspendedResourcesHolder *SuspendedResourcesHolder
			)

			if txInfo == nil {
				panic(errors.New("transactionInfo does not exist"))
			}
			method := proxy.GetMethod(serviceName,methodName)

			inNum := len(in)
			invCtx := &context2.RootContext{Context: context.Background()}
			for i := 0; i < inNum; i++ {
				if in[i].Type().String() == "context.Context" {
					if !in[i].IsNil() {
						// the user declared context as method's parameter
						invCtx =  &context2.RootContext{Context:in[i].Interface().(context.Context)}
					}
				}
				args = append(args,in[i].Interface())
			}

			tx := GetCurrentOrCreate(invCtx)
			defer tx.Resume(suspendedResourcesHolder,invCtx)

			switch txInfo.Propagation {
			case NOT_SUPPORTED:
				suspendedResourcesHolder,_ = tx.Suspend(true,invCtx)
				returnValues = proxy.Invoke(method,invCtx,serviceName,methodName,args)
				return returnValues
			case REQUIRES_NEW:
				suspendedResourcesHolder,_ = tx.Suspend(true,invCtx)
				break
			case SUPPORTS:
				if !invCtx.InGlobalTransaction() {
					returnValues = proxy.Invoke(method,invCtx,serviceName,methodName,args)
					return returnValues
				}
				break
			case REQUIRED:
				break
			case NEVER:
				if invCtx.InGlobalTransaction() {
					return proxy.ReturnWithError(method,errors.Errorf("Existing transaction found for transaction marked with propagation 'never',xid = %s",invCtx.GetXID()))
				} else {
					returnValues = proxy.Invoke(method,invCtx,serviceName,methodName,args)
					return returnValues
				}
			case MANDATORY:
				if !invCtx.InGlobalTransaction() {
					return proxy.ReturnWithError(method,errors.New("No existing transaction found for transaction marked with propagation 'mandatory'"))
				}
				break
			default:
				return proxy.ReturnWithError(method,errors.Errorf("Not Supported Propagation: %s",txInfo.Propagation.String()))
			}

			beginErr := tx.BeginWithTimeoutAndName(txInfo.TimeOut,txInfo.Name,invCtx)
			if beginErr != nil {
				return proxy.ReturnWithError(method, errors.WithStack(beginErr))
			}

			returnValues = proxy.Invoke(method,invCtx,serviceName,methodName,args)

			errValue := returnValues[len(returnValues)-1]

			//todo 只要出错就回滚，未来可以优化一下，某些错误才回滚，某些错误的情况下，可以提交
			if errValue.IsValid() && !errValue.IsNil() {
				rollbackErr := tx.Rollback(invCtx)
				if rollbackErr != nil {
					return proxy.ReturnWithError(method,errors.WithStack(rollbackErr))
				}
				return proxy.ReturnWithError(method,errors.New("rollback failure"))
			}

			commitErr := tx.Commit(invCtx)
			if commitErr != nil {
				return proxy.ReturnWithError(method,errors.WithStack(commitErr))
			}

			return returnValues
		}
	}

	numField := valueOfElem.NumField()
	for i := 0; i < numField; i++ {
		t := typeOf.Field(i)
		methodName := t.Name
		f := valueOfElem.Field(i)
		if f.Kind() == reflect.Func && f.IsValid() && f.CanSet() {
			outNum := t.Type.NumOut()

			// The latest return type of the method must be error.
			if returnType := t.Type.Out(outNum - 1); returnType != typError {
				logging.Logger.Warnf("the latest return type %s of method %q is not error", returnType, t.Name)
				continue
			}

			// do method proxy here:
			f.Set(reflect.MakeFunc(f.Type(), makeCallProxy(serviceName,methodName,v.GetMethodTransactionInfo(methodName))))
			logging.Logger.Debugf("set method [%s]", methodName)
		}
	}
}

