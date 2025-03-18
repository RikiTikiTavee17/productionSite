const initializer = {
    isInitialized: false,
    tasksInitialized: false,

    init() {
        if (this.isInitialized) return;
        this.isInitialized = true;
        console.log('Инициализация приложения');

        const { NoteV1Client } = require('./note_grpc_web_pb');
        const { CreatePersonReqest, LogInPersonRequest, CreateRequest, GetRequest, ListRequest, UpdateRequest, DeleteRequest, NoteInfo } = require('./note_pb');

        const client = new NoteV1Client('http://localhost:8080', null, null);

        const logGrpcError = (err) => {
            console.error('gRPC Error:', { message: err.message, code: err.code, details: err.details || 'No details' });
            return err.message || 'Неизвестная ошибка';
        };

        const showMessage = (element, text, color, duration = 2000) => {
            element.textContent = text;
            element.style.color = color;
            element.classList.add('fade-in');
            setTimeout(() => {
                element.textContent = '';
                element.classList.remove('fade-in');
            }, duration);
        };

        const initializeAuth = () => {
            const loginForm = document.getElementById('loginForm');
            const registerForm = document.getElementById('registerForm');
            const message = document.getElementById('message');

            if (!loginForm || !registerForm || !message) {
                console.warn('Страница авторизации не полностью загружена');
                return;
            }

            const handleFormSubmit = (form, requestBuilder, successMsg, errorMap) => {
                if (form.hasEventListener) return;
                form.addEventListener('submit', (e) => {
                    e.preventDefault();
                    const loginInput = document.getElementById(form === loginForm ? 'login' : 'regLogin');
                    const passwordInput = document.getElementById(form === loginForm ? 'password' : 'regPassword');

                    console.log('loginInput:', loginInput);
                    console.log('passwordInput:', passwordInput);

                    if (!loginInput || !passwordInput) {
                        showMessage(message, 'Ошибка: поля формы не найдены', 'red');
                        console.error('Не найдены поля:', { loginInput, passwordInput });
                        return;
                    }

                    const login = loginInput.value.trim();
                    const password = passwordInput.value.trim();

                    if (!login || !password) {
                        showMessage(message, !login ? 'Введите логин' : 'Введите пароль', 'red');
                        return;
                    }

                    const request = requestBuilder(login, password);
                    client[form === loginForm ? 'logInPerson' : 'createPerson'](request, {}, (err, response) => {
                        if (err) {
                            const errorMsg = logGrpcError(err);
                            showMessage(message, errorMap[errorMsg] || `Ошибка: ${errorMsg}`, 'red');
                        } else {
                            const userId = response.getId();
                            console.log(`${successMsg}, id:`, userId);
                            if (form === loginForm) localStorage.setItem('userId', userId);
                            showMessage(message, form === loginForm ? 'Вход успешен! Перенаправление...' : 'Регистрация успешна! Войдите.', 'green');
                            form.reset();
                            if (form === loginForm) setTimeout(() => window.location.href = '/tasks.html', 1000);
                        }
                    });
                });
                form.hasEventListener = true;
            };

            handleFormSubmit(
                loginForm,
                (login, password) => {
                    const req = new LogInPersonRequest();
                    req.setLogin(login);
                    req.setPassword(password);
                    return req;
                },
                'Успешный вход',
                { 'incorrect login or password': 'Неверный логин или пароль' }
            );

            handleFormSubmit(
                registerForm,
                (login, password) => {
                    const req = new CreatePersonReqest();
                    req.setLogin(login);
                    req.setPassword(password);
                    return req;
                },
                'Успешная регистрация',
                { 'user with this name is already registered': 'Пользователь уже существует' }
            );
        };

        const initializeTasks = () => {
            if (initializer.tasksInitialized) {
                console.log('Попытка повторной инициализации задач, пропускаем');
                return;
            }
            initializer.tasksInitialized = true;
            console.log('Инициализация задач');

            const userId = localStorage.getItem('userId');
            if (!userId) {
                window.location.href = '/index.html';
                return;
            }

            const elements = {
                taskList: document.getElementById('taskList'),
                createTaskForm: document.getElementById('createTaskForm'),
                editTaskForm: document.getElementById('editTaskForm'),
                logoutBtn: document.getElementById('logout'),
                cancelEditBtn: document.getElementById('cancelEdit'),
                taskMessage: document.getElementById('taskMessage') || document.createElement('p'),
                addTaskBtn: document.getElementById('addTaskBtn'),
                cancelCreateBtn: document.getElementById('cancelCreateBtn'),
                taskModal: document.getElementById('taskModal'),
                modalTitle: document.getElementById('modalTitle'),
                modalContent: document.getElementById('modalContent'),
                modalDeadline: document.getElementById('modalDeadline'),
                modalStatus: document.getElementById('modalStatus'),
                closeModalBtn: document.getElementById('closeModalBtn'),
                createTaskCard: document.getElementById('createTaskCard'),
                editTaskModal: document.getElementById('editTaskModal')
            };

            const missingElements = Object.entries(elements)
                .filter(([key, value]) => !value)
                .map(([key]) => key);
            if (missingElements.length > 0) {
                console.warn('Страница задач не полностью загружена. Отсутствуют элементы:', missingElements);
                return;
            }

            elements.taskMessage.id = 'taskMessage';
            elements.taskList.before(elements.taskMessage);

            const loadTasks = () => {
                const request = new ListRequest();
                request.setPersonid(parseInt(userId));

                client.list(request, {}, (err, response) => {
                    if (err) {
                        logGrpcError(err);
                        elements.taskList.innerHTML = '<p class="text-danger text-center">Ошибка загрузки задач</p>';
                        return;
                    }
                    elements.taskList.innerHTML = '';
                    const notes = response.getNotesList();
                    if (notes.length === 0) {
                        elements.taskList.innerHTML = '<p class="text-muted text-center">Нет задач</p>';
                    }
                    notes.forEach(note => {
                        const info = note.getInfo();
                        const deadline = info.getDeadLine() ? new Date(info.getDeadLine().toDate()) : null;
                        const deadlineStr = deadline ? deadline.toLocaleString() : 'Нет дедлайна';
                        const isOverdue = deadline && !info.getStatus() && deadline < new Date();
                        const deadlineClass = isOverdue ? 'text-danger' : 'text-muted';
                        const card = document.createElement('div');
                        card.className = 'col-md-4';
                        card.innerHTML = `
                            <div class="card task-card shadow-sm" data-id="${note.getId()}">
                                <div class="card-body">
                                    <h5 class="card-title">${info.getTitle() || 'Без заголовка'}</h5>
                                    <p class="card-text">${info.getContent() || 'Без описания'}</p>
                                    <p class="${deadlineClass}"><i class="fas fa-calendar-alt me-2"></i>Дедлайн: ${deadlineStr}</p>
                                    <p class="${info.getStatus() ? 'text-success' : 'text-warning'} status-text">
                                        <i class="fas ${info.getStatus() ? 'fa-check-circle' : 'fa-exclamation-circle'} me-2"></i>
                                        Статус: ${info.getStatus() ? 'Выполнено' : 'Не выполнено'}
                                    </p>
                                    <button class="btn btn-info view-btn me-2" data-id="${note.getId()}"><i class="fas fa-eye me-2"></i>Просмотр</button>
                                    <button class="btn btn-warning edit-btn me-2" data-id="${note.getId()}"><i class="fas fa-edit me-2"></i>Редактировать</button>
                                    <button class="btn btn-danger delete-btn me-2" data-id="${note.getId()}"><i class="fas fa-trash me-2"></i>Удалить</button>
                                    <button class="btn btn-success complete-btn me-2" data-id="${note.getId()}"><i class="fas fa-check me-2"></i>Выполнено</button>
                                </div>
                            </div>
                        `;
                        elements.taskList.appendChild(card);
                    });

                    console.log('Привязка обработчиков к кнопкам');
                    document.querySelectorAll('.view-btn').forEach(btn => {
                        btn.addEventListener('click', () => {
                            console.log('Нажата кнопка "Просмотр" для id:', btn.dataset.id);
                            viewTask(btn.dataset.id);
                        });
                    });
                    document.querySelectorAll('.edit-btn').forEach(btn => btn.addEventListener('click', () => editTask(btn.dataset.id)));
                    document.querySelectorAll('.delete-btn').forEach(btn => btn.addEventListener('click', () => deleteTask(btn.dataset.id)));
                    document.querySelectorAll('.complete-btn').forEach(btn => btn.addEventListener('click', () => completeTask(btn.dataset.id)));
                });
            };

            const viewTask = (id) => {
                console.log('Вызов viewTask для id:', id);
                const request = new GetRequest();
                request.setId(parseInt(id));
                client.get(request, {}, (err, response) => {
                    if (err) {
                        console.error('Ошибка в Get:', err);
                        showMessage(elements.taskMessage, `Ошибка получения задачи: ${logGrpcError(err)}`, 'red');
                        return;
                    }
                    console.log('Получен ответ от сервера:', response.getNote());
                    const note = response.getNote();
                    const info = note.getInfo();
                    elements.modalTitle.textContent = info.getTitle() || 'Без заголовка';
                    elements.modalContent.textContent = info.getContent() || 'Без описания';
                    elements.modalDeadline.textContent = info.getDeadLine() ? info.getDeadLine().toDate().toLocaleString() : 'Нет дедлайна';
                    elements.modalStatus.textContent = info.getStatus() ? 'Выполнено' : 'Не выполнено';
                    console.log('Показываем модальное окно');
                    const modal = new bootstrap.Modal(elements.taskModal);
                    modal.show();
                });
            };

            const editTask = (id) => {
                console.log('Вызов editTask для id:', id);
                const request = new GetRequest();
                request.setId(parseInt(id));
                client.get(request, {}, (err, response) => {
                    if (err) {
                        showMessage(elements.taskMessage, `Ошибка получения задачи: ${logGrpcError(err)}`, 'red');
                        return;
                    }
                    const note = response.getNote();
                    const info = note.getInfo();
                    document.getElementById('editId').value = id;
                    document.getElementById('editTitle').value = info.getTitle() || '';
                    document.getElementById('editContent').value = info.getContent() || '';
                    document.getElementById('editDeadline').value = info.getDeadLine() ? info.getDeadLine().toDate().toISOString().slice(0, 16) : '';
                    document.getElementById('editStatus').value = info.getStatus().toString();
                    const modal = new bootstrap.Modal(elements.editTaskModal);
                    modal.show();
                });
            };

            const deleteTask = (id) => {
                const request = new DeleteRequest();
                request.setId(parseInt(id));
                client.delete(request, {}, (err) => {
                    if (err) showMessage(elements.taskMessage, `Ошибка удаления: ${logGrpcError(err)}`, 'red');
                    else {
                        showMessage(elements.taskMessage, 'Задача удалена!', 'green');
                        loadTasks();
                    }
                });
            };

            const completeTask = (id) => {
                const request = new UpdateRequest();
                request.setId(parseInt(id));
                const info = new proto.note_v1.UpdateNoteInfo();
                info.setStatus(new proto.google.protobuf.BoolValue().setValue(true));
                request.setInfo(info);
                client.update(request, {}, (err) => {
                    if (err) {
                        showMessage(elements.taskMessage, `Ошибка выполнения: ${logGrpcError(err)}`, 'red');
                    } else {
                        showMessage(elements.taskMessage, 'Задача выполнена!', 'green');
                        const card = document.querySelector(`.task-card[data-id="${id}"]`);
                        if (card) {
                            const statusText = card.querySelector('.status-text');
                            statusText.className = 'status-text text-success';
                            statusText.innerHTML = '<i class="fas fa-check-circle me-2"></i>Статус: Выполнено';
                        }
                    }
                });
            };

            if (!elements.addTaskBtn.hasEventListener) {
                elements.addTaskBtn.addEventListener('click', () => {
                    console.log('Нажата кнопка "Добавить задачу"');
                    const modal = new bootstrap.Modal(elements.createTaskCard);
                    modal.show();
                });
                elements.addTaskBtn.hasEventListener = true;
            }

            if (!elements.cancelCreateBtn.hasEventListener) {
                elements.cancelCreateBtn.addEventListener('click', () => {
                    const modal = bootstrap.Modal.getInstance(elements.createTaskCard);
                    modal.hide();
                    elements.createTaskForm.reset();
                    elements.taskMessage.textContent = '';
                });
                elements.cancelCreateBtn.hasEventListener = true;
            }

            if (!elements.createTaskForm.hasEventListener) {
                elements.createTaskForm.addEventListener('submit', (e) => {
                    e.preventDefault();
                    const title = document.getElementById('title').value.trim();
                    const content = document.getElementById('content').value.trim();
                    const deadlineValue = document.getElementById('deadline').value;

                    if (!title || !content) {
                        showMessage(elements.taskMessage, 'Введите заголовок и описание', 'red');
                        return;
                    }

                    const info = new NoteInfo();
                    info.setTitle(title);
                    info.setContent(content);
                    info.setAuthor(parseInt(userId));
                    if (deadlineValue) info.setDeadLine(new proto.google.protobuf.Timestamp.fromDate(new Date(deadlineValue)));
                    info.setStatus(false);

                    const request = new CreateRequest();
                    request.setInfo(info);

                    client.create(request, {}, (err) => {
                        if (err) showMessage(elements.taskMessage, `Ошибка создания: ${logGrpcError(err)}`, 'red');
                        else {
                            showMessage(elements.taskMessage, 'Задача создана!', 'green');
                            loadTasks();
                            elements.createTaskForm.reset();
                            const modal = bootstrap.Modal.getInstance(elements.createTaskCard);
                            modal.hide();
                        }
                    });
                });
                elements.createTaskForm.hasEventListener = true;
            }

            if (!elements.editTaskForm.hasEventListener) {
                elements.editTaskForm.addEventListener('submit', (e) => {
                    e.preventDefault();
                    const id = document.getElementById('editId').value;
                    const title = document.getElementById('editTitle').value.trim();
                    const content = document.getElementById('editContent').value.trim();
                    const deadlineValue = document.getElementById('editDeadline').value;
                    const status = document.getElementById('editStatus').value === 'true';

                    if (!title || !content) {
                        showMessage(elements.taskMessage, 'Введите заголовок и описание', 'red');
                        return;
                    }

                    const info = new proto.note_v1.UpdateNoteInfo();
                    info.setTitle(new proto.google.protobuf.StringValue().setValue(title));
                    info.setContent(new proto.google.protobuf.StringValue().setValue(content));
                    if (deadlineValue) info.setDeadLine(new proto.google.protobuf.Timestamp.fromDate(new Date(deadlineValue)));
                    info.setStatus(new proto.google.protobuf.BoolValue().setValue(status));

                    const request = new UpdateRequest();
                    request.setId(parseInt(id));
                    request.setInfo(info);

                    client.update(request, {}, (err) => {
                        if (err) showMessage(elements.taskMessage, `Ошибка обновления: ${logGrpcError(err)}`, 'red');
                        else {
                            showMessage(elements.taskMessage, 'Задача обновлена!', 'green');
                            loadTasks();
                            const modal = bootstrap.Modal.getInstance(elements.editTaskModal);
                            modal.hide();
                        }
                    });
                });
                elements.editTaskForm.hasEventListener = true;
            }

            if (!elements.cancelEditBtn.hasEventListener) {
                elements.cancelEditBtn.addEventListener('click', () => {
                    const modal = bootstrap.Modal.getInstance(elements.editTaskModal);
                    modal.hide();
                    elements.taskMessage.textContent = '';
                });
                elements.cancelEditBtn.hasEventListener = true;
            }

            if (!elements.closeModalBtn.hasEventListener) {
                elements.closeModalBtn.addEventListener('click', () => {
                    const modal = bootstrap.Modal.getInstance(elements.taskModal);
                    modal.hide();
                });
                elements.closeModalBtn.hasEventListener = true;
            }

            if (!elements.logoutBtn.hasEventListener) {
                elements.logoutBtn.addEventListener('click', () => {
                    localStorage.removeItem('userId');
                    window.location.href = '/index.html';
                });
                elements.logoutBtn.hasEventListener = true;
            }

            loadTasks();
        };

        document.addEventListener('DOMContentLoaded', () => {
            console.log('Событие DOMContentLoaded сработало');
            if (document.getElementById('loginForm')) initializeAuth();
            else if (document.getElementById('taskList')) initializeTasks();
        });
    }
};

initializer.init();