package taskRunManager

type Task func()

type Manager interface {
	Run(t Task)
	Wait()
}
