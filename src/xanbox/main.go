package main

import (
	"log"
	"net/http"

	"github.com/Kooperativerupestre/XanboX/src/xanbox/database"
	"github.com/Kooperativerupestre/XanboX/src/xanbox/task"
	"github.com/Kooperativerupestre/XanboX/src/xanbox/user"
	"github.com/go-chi/chi/v5"
)

func main() {
	pool := database.NewPool()

	r := chi.NewRouter()

	userRepository := user.NewUserRepository(pool)
	userService := user.NewUserService(userRepository)
	userRouter := user.NewUserRouter(userService)

	r.Mount("/users", userRouter)

	taskRepository := task.NewTaskRepository(pool)
	taskService := task.NewTaskService(taskRepository)
	taskRouter := task.NewTaskRouter(taskService)

	r.Mount("/tasks", taskRouter)

	log.Fatal(http.ListenAndServe(":8080", r))
}
