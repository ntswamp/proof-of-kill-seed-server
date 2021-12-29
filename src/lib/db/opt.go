package lib_db

type Task struct {
	F    func(...interface{}) error
	Args []interface{}
}

/**
 * 保存するときのオプション.
 */
type SaveOptions struct {
	Fields      []string
	SavedTask   *Task
	ForceInsert bool
	ForceUpdate bool
}

var OptInsert = &SaveOptions{
	ForceInsert: true,
}

/**
 * 保存予約情報.
 */
type ReservedInfo struct {
	Model       interface{}
	Fields      []string
	fieldMap    map[string]bool
	ForceInsert bool
	ForceUpdate bool
	SavedTasks  []*Task
}
