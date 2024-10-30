package main

import (
	// "fmt"
	"net/http"
	// "encoding/json"
	// "time"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Fprintf(w, "status: available\n")
	// fmt.Fprintf(w, "environment: %s\n", app.config.env)
	// fmt.Fprintf(w, "version: %s\n", version)

	// js := `{"status": "available", "environment": %q, "version": %q}`
	// js = fmt.Sprintf(js, app.config.env, version)

	// w.Header().Set("Content-Type", "application/json")

	// w.Write([]byte(js))

	// data := map[string]string{
	// 	"status": "available",
	// 	"environment": app.config.env,
	// 	"version": version,
	// }

	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	// time.Sleep(4*time.Second)

	// js, err := json.Marshal(data)
	err := app.WriteJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		// return
	}

	// js = append(js, '\n')

	// w.Header().Set("Content-Type", "application/json")

	// w.Write(js)
}
