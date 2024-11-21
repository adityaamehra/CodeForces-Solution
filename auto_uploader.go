package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	gitHubAPIURL    = "https://api.github.com"
	gitHubRepoOwner = "adityaamehra"                             // Replace with your GitHub username
	gitHubRepoName  = "CodeForces-Solution"                      // Replace with your repository name
	gitHubToken     = "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" // Replace with your GitHub token
)

func main() {
	handle := "Adityaa_Mehra"
	submissions := getSubmissions(handle)
	acceptedProblems := make([]AcceptedProblem, 0, len(submissions))
	for _, v := range submissions {
		if v.Verdict == "OK" {
			acceptedProblem := AcceptedProblem{v.ContestID, v.ID, v.Problem.Name, v.ProgrammingLanguage, v.Problem.Index}
			acceptedProblems = append(acceptedProblems, acceptedProblem)
		}
	}

	for _, acceptedProblem := range acceptedProblems {
		fileName := acceptedProblem.getFileName()
		fileContent := acceptedProblem.getLink()
		exists, err := checkFileExistsInGitHubRepo(fileName)
		if err != nil {
			panic(err)
		}
		if !exists {
			err = createFileInGitHubRepo(fileName, fileContent)
			if err != nil {
				panic(err)
			}
			fmt.Println("Created File:", fileName)
		} else {
			fmt.Println("File already exists:", fileName)
		}
	}
	fmt.Println("All files have been processed.")
}

func getSubmissions(handle string) Submissions {
	resp, err := http.Get("http://codeforces.com/api/user.status?handle=" + handle + "&from=1&count=10000")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	arr := status{}
	json.Unmarshal(body, &arr)
	if arr.Status == "FAILED" {
		fmt.Println("Codeforces handle incorrect! Please try again.")
		os.Exit(1)
	}
	return arr.Result
}

func checkFileExistsInGitHubRepo(fileName string) (bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", gitHubAPIURL, gitHubRepoOwner, gitHubRepoName, url.PathEscape(fileName))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "token "+gitHubToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true, nil
	} else if resp.StatusCode == 404 {
		return false, nil
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		return false, fmt.Errorf("GitHub API error: %s", body)
	}
}

func createFileInGitHubRepo(fileName, fileContent string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", gitHubAPIURL, gitHubRepoOwner, gitHubRepoName, url.PathEscape(fileName))
	fileContentEncoded := base64.StdEncoding.EncodeToString([]byte(fileContent))
	requestBody := map[string]string{
		"message": "Add " + fileName,
		"content": fileContentEncoded,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+gitHubToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %s", body)
	}
	return nil
}

type Submissions []struct {
	ID                  int   `json:"id"`
	ContestID           int   `json:"contestId"`
	CreationTimeSeconds int   `json:"creationTimeSeconds"`
	RelativeTimeSeconds int64 `json:"relativeTimeSeconds"`
	Problem             struct {
		ContestID int      `json:"contestId"`
		Index     string   `json:"index"`
		Name      string   `json:"name"`
		Type      string   `json:"type"`
		Tags      []string `json:"tags"`
	} `json:"problem"`
	Author struct {
		ContestID int `json:"contestId"`
		Members   []struct {
			Handle string `json:"handle"`
		} `json:"members"`
		ParticipantType  string `json:"participantType"`
		Ghost            bool   `json:"ghost"`
		StartTimeSeconds int    `json:"startTimeSeconds"`
	} `json:"author"`
	ProgrammingLanguage string `json:"programmingLanguage"`
	Verdict             string `json:"verdict"`
	Testset             string `json:"testset"`
	PassedTestCount     int    `json:"passedTestCount"`
	TimeConsumedMillis  int    `json:"timeConsumedMillis"`
	MemoryConsumedBytes int    `json:"memoryConsumedBytes"`
}
type status struct {
	Status string      `json:"status"`
	Result Submissions `json:"result"`
}

type AcceptedProblem struct {
	contestId, submissionID int
	name, language, index   string
}

func (l AcceptedProblem) getLink() string {
	var gymOrContest string
	if l.contestId >= 10000 {
		gymOrContest = "gym"
	} else {
		gymOrContest = "contest"
	}
	return "https://codeforces.com/" + gymOrContest + "/" + strconv.Itoa(l.contestId) + "/submission/" + strconv.Itoa(l.submissionID)
}

func (l AcceptedProblem) getFileName() string {
	return strconv.Itoa(l.contestId) + "-" + l.index + "_" + normalizeProblemName(l.name) + "." + normalizeLanguageName(l.language)
}

func normalizeLanguageName(s string) string {
	if strings.Contains(s, "Java") || strings.Contains(s, "java") {
		return "java"
	} else if strings.Contains(s, "Gnu") || strings.Contains(s, "C++") || strings.Contains(s, "c++") {
		return "cpp"
	} else if strings.Contains(s, "Python") || strings.Contains(s, "python") || strings.Contains(s, "PyPy 3-64") {
		return "py"
	} else if strings.Contains(s, "Go") || strings.Contains(s, "go") {
		return "go"
	} else if strings.Contains(s, "GNU C11") {
		return "c"
	} else if strings.Contains(s, "Kotlin 1.7") {
		return "kt"
	}
	return s
}

func normalizeProblemName(name string) string {
	splitted := strings.Split(name, " ")
	var res string
	for _, v := range splitted {
		res += removeSlashes(v)
	}
	return res
}

func removeSlashes(s string) string {
	var res string
	for i := 0; i < len(s); i++ {
		if s[i] != '/' {
			res += string(s[i])
		}
	}
	return res
}
