package params

const (
	ActionExecutedEventName = "ActionExecuted"         // ActionExecutedEventName is the name of the event emitted when an action is executed.
	MultiActionMethodName   = "executeMultipleActions" // MultiActionMethodName is the name of the method that executes multiple actions.
)

var Params = map[string]interface{}{
	"ActionExecutedEventName": ActionExecutedEventName,
	"MultiActionMethodName":   MultiActionMethodName,
}
