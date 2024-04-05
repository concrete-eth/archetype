package params

const (
	ActionExecutedEventName = "ActionExecuted" // ActionExecutedEventName is the name of the event emitted when an action is executed.
)

var Params = map[string]interface{}{
	"ActionExecutedEventName": ActionExecutedEventName,
}
