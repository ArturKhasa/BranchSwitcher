package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

const repoPath = "./repo" // Укажите путь к репозиторию

func getBranches() ([]string, error) {
	// Выполняем `git fetch`, чтобы обновить список веток
	fetchCmd := exec.Command("git", "fetch", "--prune")
	fetchCmd.Dir = repoPath
	if err := fetchCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to fetch branches: %v", err)
	}

	cmd := exec.Command("git", "branch", "--list")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	branches := strings.Split(string(output), "\n")
	for i, branch := range branches {
		branches[i] = strings.TrimSpace(strings.TrimPrefix(branch, "* "))
	}

	return branches, nil
}

func switchBranch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	branch := r.FormValue("branch")
	if branch == "" {
		http.Error(w, "Branch name is required", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to switch branch: %s", output), http.StatusInternalServerError)
		return
	}

	deployCmd := exec.Command("sh", ".scripts/deploy.sh")
	deployCmd.Dir = repoPath
	deployOutput, deployErr := deployCmd.CombinedOutput()
	if deployErr != nil {
		http.Error(w, fmt.Sprintf("Failed to run deploy script: %s", deployOutput), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Switched to branch: %s\nDeploy script executed successfully", branch)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	branches, err := getBranches()
	if err != nil {
		http.Error(w, "Failed to retrieve branches", http.StatusInternalServerError)
		return
	}

	html := `<html>
	<head><title>Branch Switcher</title></head>
	<body>
		<h1>Switch Git Branch</h1>
		<form action="/switch" method="POST">
			<select name="branch" required>
			` + func() string {
		var options string
		for _, branch := range branches {
			if branch != "" {
				options += fmt.Sprintf("<option value=\"%s\">%s</option>", branch, branch)
			}
		}
		return options
	}() + `
			</select>
			<button type="submit">Switch</button>
		</form>
	</body>
	</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func main() {
	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/switch", switchBranch)

	port := ":8080"
	fmt.Printf("Server running at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// Dockerfile
//FROM golang:1.20
//WORKDIR /app
//COPY . .
//RUN go build -o git-branch-switcher
//CMD ["./git-branch-switcher"]
//EXPOSE 8080
