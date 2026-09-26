package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/Kooperativerupestre/XanboX/src/xanbox/database"
	"github.com/Kooperativerupestre/XanboX/src/xanbox/task"
	"github.com/Kooperativerupestre/XanboX/src/xanbox/task/execution"
	"github.com/Kooperativerupestre/XanboX/src/xanbox/task/source"
	"github.com/Kooperativerupestre/XanboX/src/xanbox/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/moby/moby/client"
	"github.com/uptrace/bun"
)

func getTestDatabaseDSN(t *testing.T) string {
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load("../.env")

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL environment variable is required and must not be empty")
	}

	u, err := url.Parse(dsn)
	if err == nil && u.Hostname() != "" {
		if _, lookupErr := net.LookupHost(u.Hostname()); lookupErr != nil {
			dockerCli, err := client.NewClientWithOpts(client.FromEnv)
			if err == nil {
				defer dockerCli.Close()
				for _, name := range []string{"xanbox-test-postgres-1", "test-postgres"} {
					inspect, err := dockerCli.ContainerInspect(context.Background(), name, client.ContainerInspectOptions{})
					if err == nil && inspect.Container.NetworkSettings != nil {
						for _, netConfig := range inspect.Container.NetworkSettings.Networks {
							if netConfig.IPAddress.IsValid() {
								port := u.Port()
								if port == "" {
									port = "5432"
								}
								u.Host = net.JoinHostPort(netConfig.IPAddress.String(), port)
								return u.String()
							}
						}
					}
				}
			}
		}
	}

	return dsn
}

func setupTestEnvironment(t *testing.T) (*bun.DB, *httptest.Server, *client.Client) {
	testDSN := getTestDatabaseDSN(t)
	db := database.ConnectDB(testDSN)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var pingRes int
	if err := db.NewSelect().ColumnExpr("1").Scan(ctx, &pingRes); err != nil {
		t.Fatalf("failed to connect to test database (%s): %v", testDSN, err)
	}

	dockerCli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Fatalf("failed to initialize docker client: %v", err)
	}

	ts := execution.NewTaskExecutionsStorage()
	tm, err := execution.NewTaskManager(ts)
	if err != nil {
		t.Fatalf("failed to initialize task manager: %v", err)
	}

	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo)
	userRouter := user.NewUserRouter(userService)

	taskRepo := task.NewTaskRepository(db)
	taskService := task.NewTaskService(taskRepo, tm)
	taskRouter := task.NewTaskRouter(taskService)

	r := chi.NewRouter()
	r.Mount("/users", userRouter)
	r.Mount("/tasks", taskRouter)

	server := httptest.NewServer(r)

	t.Cleanup(func() {
		server.Close()
		dockerCli.Close()
		db.Close()
	})

	return db, server, dockerCli
}

func createTestUser(t *testing.T, db *bun.DB) uuid.UUID {
	id := uuid.New()
	u := &user.User{
		ID:   id,
		Name: fmt.Sprintf("test-user-%d", time.Now().UnixNano()),
	}

	_, err := db.NewInsert().Model(u).Exec(context.Background())
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.NewDelete().Model((*user.User)(nil)).Where("id = ?", id).Exec(context.Background())
	})

	return id
}

func TestSync(t *testing.T) {
	db, server, dockerCli := setupTestEnvironment(t)
	userID := createTestUser(t, db)

	taskPayload := task.CreateTaskRequest{
		Image:                  "python:3.13",
		EnvironmentPrepareCode: []string{"echo 'setup completed'"},
		Source: []source.File{
			{
				Path: "main.py",
				Code: "print('hello e2e')\n",
			},
		},
		ExecutionCode: "python3 main.py",
		Maker:         userID,
	}

	bodyBytes, err := json.Marshal(taskPayload)
	if err != nil {
		t.Fatalf("failed to marshal task payload: %v", err)
	}

	createResp, err := http.Post(server.URL+"/tasks", "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(createResp.Body)
		t.Fatalf("POST /tasks returned status %d: %s", createResp.StatusCode, string(b))
	}

	var createOutput struct {
		ID       uuid.UUID `json:"id"`
		DockerID string    `json:"docker_id"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&createOutput); err != nil {
		t.Fatalf("failed to decode task response: %v", err)
	}

	taskID := createOutput.ID

	// Verify task is initialised as pending
	getResp, err := http.Get(fmt.Sprintf("%s/tasks/%s", server.URL, taskID.String()))
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	defer getResp.Body.Close()

	var taskObj task.Task
	if err := json.NewDecoder(getResp.Body).Decode(&taskObj); err != nil {
		t.Fatalf("failed to decode task: %v", err)
	}
	if taskObj.Status != task.TaskPending {
		t.Fatalf("expected initial status to be pending, got %s", taskObj.Status)
	}

	// Poll sync until task finishes and is marked successful
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var finalStatus task.TaskStatus
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timed out waiting for task to sync and reach successful status")
		default:
		}

		syncResp, err := http.Post(fmt.Sprintf("%s/tasks/%s/sync", server.URL, taskID.String()), "application/json", nil)
		if err != nil {
			t.Fatalf("failed to call sync endpoint: %v", err)
		}
		if syncResp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(syncResp.Body)
			syncResp.Body.Close()
			t.Fatalf("POST /tasks/%s/sync returned status %d: %s", taskID, syncResp.StatusCode, string(b))
		}
		syncResp.Body.Close()

		checkResp, err := http.Get(fmt.Sprintf("%s/tasks/%s", server.URL, taskID.String()))
		if err != nil {
			t.Fatalf("failed to check task after sync: %v", err)
		}

		var currentTask task.Task
		_ = json.NewDecoder(checkResp.Body).Decode(&currentTask)
		checkResp.Body.Close()

		if currentTask.Status != task.TaskPending {
			finalStatus = currentTask.Status
			break
		}

		time.Sleep(300 * time.Millisecond)
	}

	if finalStatus != task.TaskSuccessful {
		t.Fatalf("expected task status to be %s, got %s", task.TaskSuccessful, finalStatus)
	}

	// Verify idempotency of sync on an already completed task
	repeatSyncResp, err := http.Post(fmt.Sprintf("%s/tasks/%s/sync", server.URL, taskID.String()), "application/json", nil)
	if err != nil {
		t.Fatalf("failed to call repeated sync: %v", err)
	}
	defer repeatSyncResp.Body.Close()
	if repeatSyncResp.StatusCode != http.StatusOK {
		t.Fatalf("repeated sync expected status 200 OK, got %d", repeatSyncResp.StatusCode)
	}

	// Verify database cleanup of execution records
	var linkCount int
	err = db.NewSelect().
		Table("task_executions").
		Where("task_id = ?", taskID).
		ColumnExpr("count(*)").
		Scan(context.Background(), &linkCount)
	if err != nil {
		t.Fatalf("failed to query task_executions: %v", err)
	}
	if linkCount != 0 {
		t.Fatalf("expected task_executions link to be deleted, found %d", linkCount)
	}

	// Verify container is removed from Docker
	if createOutput.DockerID != "" {
		containers, err := dockerCli.ContainerList(context.Background(), client.ContainerListOptions{All: true})
		if err == nil {
			for _, c := range containers.Items {
				if c.ID == createOutput.DockerID {
					t.Fatalf("container %s was expected to be deleted, but still exists", createOutput.DockerID)
				}
			}
		}
	}
}

