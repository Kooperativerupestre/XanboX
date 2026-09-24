package execution

import "sync"

type TaskExecutionsStorage struct {
	mu    sync.RWMutex
	tasks map[string]*executionContainer
}

func NewTaskExecutionsStorage() *TaskExecutionsStorage {
	return &TaskExecutionsStorage{
		tasks: make(map[string]*executionContainer),
	}
}

func (tr *TaskExecutionsStorage) get(id string) (*executionContainer, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()

	execution, exists := tr.tasks[id]

	return execution, exists
}

func (tr *TaskExecutionsStorage) add(execution *executionContainer) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.tasks[execution.execID] = execution
}

func (tr *TaskExecutionsStorage) delete(id string) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	delete(tr.tasks, id)
}
