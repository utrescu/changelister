package commits

import (
	"io"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/utrescu/changelister/tags"
)

type CommitData struct {
	Type      string
	Header    string
	Scope     string
	Body      string
	Author    string
	DateTime  string
	Important bool
	Group     string
}

type ChangelogData struct {
	Tag     string
	Message string
	Date    string
	Commits map[string][]CommitData
}

func getKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func ProcessTagCommits(repo *git.Repository, tags tags.TagInfo, labels map[string]string) ChangelogData {

	chIter, err := repo.Log(&git.LogOptions{
		From:  tags.Stop,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		panic(err)
	}
	defer chIter.Close()

	logs := make(map[string][]CommitData)

	for {
		commit, err := chIter.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("error iterating commits: %v", err)
		}

		// Stop in first commit
		if commit.Hash == tags.Start {
			break
		}

		messageLog, valid := ProcessMessage(commit, labels)
		if valid {
			if currentLogs, exists := logs[messageLog.Group]; !exists {
				logs[messageLog.Group] = []CommitData{messageLog}
			} else {
				currentLogs = append(currentLogs, messageLog)
				logs[messageLog.Group] = currentLogs
			}
		}
	}

	changelog := ChangelogData{
		Tag:     tags.Name,
		Message: tags.Message,
		Date:    tags.Date,
		Commits: logs,
	}
	return changelog
}

func ProcessMessage(commit *object.Commit, labels map[string]string) (CommitData, bool) {

	commitTypes := getKeys(labels)

	data, valid := ParseCommitMessage(commit.Message, commitTypes, labels)
	if valid {
		if slices.Contains(commitTypes, data.Type) {
			data.Author = commit.Author.Name
			data.DateTime = commit.Author.When.String()
			return *data, true
		}
	}
	return CommitData{}, false
}

func ParseCommitMessage(commitMessage string, commitTypes []string, labels map[string]string) (*CommitData, bool) {

	lines := strings.Split(commitMessage, "\n")

	if len(lines) == 0 {
		return nil, false
	}

	// The first line is the header
	header := lines[0]

	data, valid := SplitHeaderAndValidate(header, commitTypes)

	if !valid {
		return nil, false
	}
	// The rest is the body
	data.Body = strings.Join(lines[1:], "\n")
	data.Group = labels[data.Type]

	breaking := strings.Contains(data.Body, "BREAKING CHANGE") || strings.Contains(data.Body, "BREAKING-CHANGE")
	data.Important = data.Important || breaking

	return data, true
}

func SplitHeaderAndValidate(message string, commitTypes []string) (*CommitData, bool) {

	groupCommitTypes := "(" + strings.Join(commitTypes, "|") + ")"
	re := regexp.MustCompile(`(?m)` + groupCommitTypes + `\s*(\(.+\))?(!)?: (.*)`)

	match := re.FindStringSubmatch(message)

	if match != nil {
		data := &CommitData{
			Type:      match[1],
			Header:    strings.TrimSuffix(match[4], "\n"),
			Scope:     match[2],
			Body:      "",
			Important: match[3] == "!",
			Author:    "",
			DateTime:  "",
		}
		return data, true
	}
	return nil, false
}
