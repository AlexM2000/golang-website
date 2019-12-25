package main

import (
	_ "github.com/lib/pq" //Driver for PostgreSQL
)

/*func main() {
	r := mux.NewRouter()
	r.HandleFunc("/tasks", getTasks).Methods("GET")
	r.HandleFunc("/tasks", createTask).Methods("POST")
	r.HandleFunc("/tasks/{id}", completeTask).Methods("PUT")
	fmt.Println("Started serving port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	tasks := getTasksFromDb()
	json.NewEncoder(w).Encode(tasks)
}

func getTasksFromDb() []db.Task {
	dbase, err := Open()
	if err != nil {
		panic(err)
	}
	defer dbase.Close()
	Tasks := []db.Task{}

	rows, err := dbase.Query("SELECT * from todos where complete is null")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		t := db.Task{}
		err := rows.Scan(&t.ID, &t.Text, &t.CreateTime, &t.Complete)
		if err != nil {
			fmt.Println(err)
			continue
		}
		Tasks = append(Tasks, t)
	}
	return Tasks
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var newTask db.Task
	json.NewDecoder(r.Body).Decode(&newTask)
	task := createTaskInDb(newTask)
	json.NewEncoder(w).Encode(task)
}

func createTaskInDb(task db.Task) db.Task {
	dbase, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	_, err = dbase.Exec("insert into todos (task) values($1);", task.Text)
	if err != nil {
		log.Fatal(err)
	}
	row := dbase.QueryRow("select * from todos where date_create=(select max(date_create) from todos)")
	t := db.Task{}
	err = row.Scan(&t.ID, &t.Text, &t.CreateTime, &t.Complete)
	fmt.Println("Scan success")
	if err != nil {
		fmt.Println(err)
	}
	return t
}

func completeTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		panic(err)
	}
	var completeTask db.Task
	json.NewDecoder(r.Body).Decode(&completeTask)
	completeTask.ID = id
	completedTask := completeTaskInDb(completeTask)
	json.NewEncoder(w).Encode(completedTask)
}

func completeTaskInDb(task db.Task) db.Task {
	dbase, err := Open()
	if err != nil {
		log.Fatal(err)
	}
	defer dbase.Close()
	_, err = dbase.Exec("update todos set complete = TRUE where id = $1", task.ID)
	if err != nil {
		log.Fatal(err)
	}
	row := dbase.QueryRow("select * from todos where id = $1", task.ID)
	t := db.Task{}
	err = row.Scan(&t.ID, &t.Text, &t.CreateTime, &t.Complete)
	fmt.Println("Scan success")
	if err != nil {
		fmt.Println(err)
	}
	return t
}*/
