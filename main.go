package main

import (
	"io"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	tags "github.com/utrescu/changelister/tags"
	template "github.com/utrescu/changelister/template"
	config "github.com/utrescu/changelister/config"
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


func main() {

	config := config.LoadConfig()

	// Obir el repositori
	repo, err := git.PlainOpen(config.Path)
	if err != nil {
		log.Println(err)
		return
	}

	listTags, err := tags.ProcessTags(repo, config.Tag)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	changelogData := []ChangelogData{}

	for _, tags := range listTags {

		// Processar les dades de l'etiqueta
		processedTag := ProcessTagCommits(repo, tags, config.CommitTypes.Show)
		changelogData = append(changelogData, processedTag)

	}

	template.ProcessTemplate("changelog.tmpl", "changelog.md", changelogData)

}

func ProcessTagCommits(repo *git.Repository, tags tags.TagInfo, commitTypes []string) ChangelogData {

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
			log.Fatalf("Error iterant commits: %v", err)
		}

		// Si arribem al commit antic, parem
		if commit.Hash == tags.Start {
			break
		}

		log, valid := ProcessMessage(commit, commitTypes)
		if valid {
			if currentLogs, exists := logs[log.Group]; !exists {
				logs[log.Group] = []CommitData{log}
			} else {
				currentLogs = append(currentLogs, log)
				logs[log.Group] = currentLogs
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

func ProcessMessage(commit *object.Commit, commitTypes []string) (CommitData, bool) {
	newmessage := strings.Trim(commit.Message, " ")
	newmessage = strings.TrimSuffix(newmessage, "\n")

	data, valid := ProcessMessageAndValidate(newmessage, commitTypes)
	if valid {
		if slices.Contains(commitTypes, data.Type) {
			data.Author = commit.Author.Name
			data.DateTime = commit.Author.When.String()
			return *data, true
		}
	}
	return CommitData{}, false
}

func ProcessMessageAndValidate(message string, commitTypes []string) (*CommitData, bool) {

	groupCommitTypes := "(" + strings.Join(commitTypes, "|") + ")"
	re := regexp.MustCompile(`(?m)` + groupCommitTypes + `\s*(\(.+\))?(!)?:(.*)`)

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

		switch data.Type {
		case "doc", "feat":
			data.Group = "added"
		case "fix":
			data.Group = "fixed"
		case "refactor":
			data.Group = "changed"
		case "chore", "ci", "build":
			data.Group = "other"
		}

		return data, true
	}
	return nil, false
}
