package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/AlexJudin/go_final_project/model"
	"github.com/AlexJudin/go_final_project/usecases"
)

var messageError string

type TaskHandler struct {
	uc usecases.Task
}

func NewTaskHandler(uc usecases.Task) TaskHandler {
	return TaskHandler{uc: uc}
}

type errResponse struct {
	Error string `json:"error"`
}

func (h *TaskHandler) GetNextDate(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	nowTime, err := time.Parse(model.TimeFormat, now)
	if err != nil {
		log.Errorf("failed to parse time. Error: %+v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	date := r.FormValue("date")
	repeat := r.FormValue("repeat")

	nextDate, err := h.uc.GetNextDate(nowTime, date, repeat)
	if err != nil {
		log.Errorf("failed to get next date. Error: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(nextDate))
	if err != nil {
		log.Errorf("failed to write response. Error: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateTask ... Добавить новую задачу
// @Summary Добавить новую задачу
// @Description Добавить новую задачу
// @Accept json
// @Tags Task
// @Param Body body model.Task true "Параметры задачи"
// @Success 201 {object} model.TaskResp
// @Failure 400 {object} errResponse
// @Failure 500 {object} errResponse
// @Router /api/task [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var (
		task model.Task
		buf  bytes.Buffer
	)

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Errorf("create task error: %+v", err)
		messageError = "Переданы некорректные параметры задачи."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		log.Errorf("create task error: %+v", err)
		messageError = "Не удалось прочитать параметры задачи."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	dateTaskNow := time.Now().Format(model.TimeFormat)
	err = checkTaskRequest(&task, dateTaskNow)
	if err != nil {
		log.Errorf("create task error: %+v", err)
		// Скорректировать описание ошибки
		messageError = "Переданы некорректные данные о платежной операции."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	pastDay := dateTaskNow > task.Date

	taskResp, err := h.uc.CreateTask(&task, pastDay)
	if err != nil {
		log.Errorf("create task error: %+v", err)
		messageError = fmt.Sprintf("Ошибка сервера, не удалось сохранить задачу [%s] от [%s]. Попробуйте позже или обратитесь в тех. поддержку.", task.Title, task.Date)

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	resp, err := json.Marshal(taskResp)
	if err != nil {
		log.Errorf("create task error: %+v", err)
		messageError = "Ошибка сервера. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("create task error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusServiceUnavailable, messageError, w)
	}
}

// GetTasks ... Получить список ближайших задач
// @Summary Получить список ближайших задач
// @Description Получить список ближайших задач
// @Accept json
// @Tags Task
// @Param search query string true "Строка поиска"
// @Success 200 {object} model.TasksResp
// @Failure 400 {object} errResponse
// @Failure 500 {object} errResponse
// @Router /api/tasks [get]
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	searchString := r.FormValue("search")

	tasksResp, err := h.uc.GetTasks(searchString)
	if err != nil {
		log.Errorf("get tasks error: %+v", err)
		messageError = "Ошибка сервера, не удалось получить список задач. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	resp, err := json.Marshal(tasksResp)
	if err != nil {
		log.Errorf("get tasks error: %+v", err)
		messageError = "Ошибка сервера. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("get tasks error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusServiceUnavailable, messageError, w)
	}
}

// GetTaskById ... Получить задачу
// @Summary Получить задачу
// @Description Получить задачу
// @Accept json
// @Tags Task
// @Param id query string true "Идентификатор задачи"
// @Success 200 {object} model.TaskResp
// @Failure 400 {object} errResponse
// @Failure 500 {object} errResponse
// @Router /api/task [get]
func (h *TaskHandler) GetTaskById(w http.ResponseWriter, r *http.Request) {
	taskId := r.FormValue("id")
	if taskId == "" {
		err := fmt.Errorf("task id is empty")
		log.Errorf("get task by id error: %+v", err)
		messageError = "Не передан идентификатор, получение параметров задачи невозможно."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	taskResp, err := h.uc.GetTaskById(taskId)
	if err != nil {
		log.Errorf("get task by id error: %+v", err)
		messageError = fmt.Sprintf("Ошибка сервера, не удалось получить параметры задачи [%s]. Попробуйте позже или обратитесь в тех. поддержку.", taskId)

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	resp, err := json.Marshal(taskResp)
	if err != nil {
		log.Errorf("get task by id error: %+v", err)
		messageError = "Ошибка сервера. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resp)
	if err != nil {
		log.Errorf("get task by id error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusServiceUnavailable, messageError, w)
	}
}

// UpdateTask ... Редактировать задачу
// @Summary Редактировать задачу
// @Description Редактировать задачу
// @Accept json
// @Tags Task
// @Param Body body model.Task true "Параметры задачи"
// @Success 200 {string} string
// @Failure 400 {object} errResponse
// @Failure 500 {object} errResponse
// @Router /api/task [put]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var (
		task model.Task
		buf  bytes.Buffer
	)

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Errorf("update task error: %+v", err)
		messageError = "Переданы некорректные параметры задачи."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		log.Errorf("update task error: %+v", err)
		messageError = "Не удалось прочитать параметры задачи."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	dateTaskNow := time.Now().Format(model.TimeFormat)
	err = checkTaskRequest(&task, dateTaskNow)
	if err != nil {
		log.Errorf("update task error: %+v", err)
		// Скорректировать описание ошибки
		messageError = "Переданы некорректные данные о платежной операции."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	pastDay := dateTaskNow > task.Date

	err = h.uc.UpdateTask(&task, pastDay)
	if err != nil {
		log.Errorf("update task error: %+v", err)
		messageError = fmt.Sprintf("Ошибка сервера, не удалось обновить задачу [%s]. Попробуйте позже или обратитесь в тех. поддержку.", task.Id)

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("{}"))
	if err != nil {
		log.Errorf("update task error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusServiceUnavailable, messageError, w)
	}
}

// MakeTaskDone ... Выполнить задачу
// @Summary Выполнить задачу
// @Description Выполнить задачу
// @Accept json
// @Tags Task
// @Param id query string true "Идентификатор задачи"
// @Success 200 {object} model.TaskResp
// @Failure 400 {object} errResponse
// @Failure 500 {object} errResponse
// @Router /api/task/done [post]
func (h *TaskHandler) MakeTaskDone(w http.ResponseWriter, r *http.Request) {
	taskId := r.FormValue("id")
	if taskId == "" {
		err := fmt.Errorf("task id is empty")
		log.Errorf("make task done error: %+v", err)
		messageError = "Не передан идентификатор задачи, невозможно установить отметку выполнения."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	err := h.uc.MakeTaskDone(taskId)
	if err != nil {
		log.Errorf("make task done error: %+v", err)
		messageError = fmt.Sprintf("Ошибка сервера, не удалось обновить задачу [%s]. Попробуйте позже или обратитесь в тех. поддержку.", taskId)

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("{}"))
	if err != nil {
		log.Errorf("make task done error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusServiceUnavailable, messageError, w)
	}
}

// DeleteTask ... Удалить задачу
// @Summary Удалить задачу
// @Description Удалить задачу
// @Accept json
// @Tags Task
// @Param id query string true "Идентификатор задачи"
// @Success 200 {object} model.TaskResp
// @Failure 400 {object} errResponse
// @Failure 500 {object} errResponse
// @Router /api/task [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.FormValue("id")
	if taskId == "" {
		err := fmt.Errorf("task id is empty")
		log.Errorf("delete task error: %+v", err)
		messageError = "Не передан идентификатор, невозможно удалить задачу."

		returnErr(http.StatusBadRequest, messageError, w)
		return
	}

	err := h.uc.DeleteTask(taskId)
	if err != nil {
		log.Errorf("delete task error: %+v", err)
		messageError = fmt.Sprintf("Ошибка сервера, не удалось удалить задачу [%s]. Попробуйте позже или обратитесь в тех. поддержку.", taskId)

		returnErr(http.StatusInternalServerError, messageError, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("{}"))
	if err != nil {
		log.Errorf("delete task error: %+v", err)
		messageError = "Сервер недоступен. Попробуйте позже или обратитесь в тех. поддержку."

		returnErr(http.StatusServiceUnavailable, messageError, w)
	}
}

func checkTaskRequest(task *model.Task, dateTaskNow string) error {
	if task.Title == "" {
		return fmt.Errorf("task title is empty")
	}

	if task.Date == "" {
		task.Date = dateTaskNow
		return nil
	}

	_, err := time.Parse(model.TimeFormat, task.Date)
	if err != nil {
		return fmt.Errorf("task date is invalid")
	}

	if task.Date < dateTaskNow && task.Repeat == "" {
		task.Date = dateTaskNow
	}

	return nil
}

func returnErr(status int, messageError string, w http.ResponseWriter) {
	message := errResponse{
		Error: messageError,
	}

	messageJson, err := json.Marshal(message)
	if err != nil {
		status = http.StatusInternalServerError
		messageJson = []byte("{\"error\":\"" + err.Error() + "\"}")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(messageJson)
	if err != nil {
		log.Errorf("get wallet balance by UUID error: %+v", err)
	}
}
