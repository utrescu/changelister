package main

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5"
	tags "github.com/utrescu/changelister/tags"
	template "github.com/utrescu/changelister/template"
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
	Commits map[string][]CommitData
}

var commitTypes = []string{"feat", "fix", "chore", "docs", "doc", "style", "refactor", "perf", "test", "build", "ci", "revert"}

func main() {

	path := "/home/xavier/work-institut/0-manteniment/Manteniment-Aules"
	tag := "curs25-26"
	commitTypesToShow := []string{"feat", "refactor", "docs", "doc"}

	// Obir el repositori
	repo, err := git.PlainOpen(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	listTags, err := tags.ProcessTags(repo, tag)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	changelogData := []ChangelogData{}

	for _, tags := range listTags {

		// Processar les dades de l'etiqueta
		processedTag := ProcessTagCommits(repo, tags, commitTypesToShow)
		changelogData = append(changelogData, processedTag)

	}

	template.ProcessTemplate("changelog.tmpl", "changelog.md", changelogData)

}

func ProcessTagCommits(repo *git.Repository, tags tags.TagInfo, commitTypesToShow []string) ChangelogData {

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

		log, valid := ProcessMessage(commit.Message, commitTypesToShow)
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
		Commits: logs,
	}
	return changelog
}

func ProcessMessage(message string, commitTypesToShow []string) (CommitData, bool) {
	newmessage := strings.Trim(message, " ")
	newmessage = strings.TrimSuffix(newmessage, "\n")

	data, valid := ProcessMessageAndValidate(newmessage)
	if valid {
		if slices.Contains(commitTypesToShow, data.Type) {
			return *data, true
		}
	}
	return CommitData{}, false
}



func ProcessMessageAndValidate(message string) (*CommitData, bool) {

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
		}

		switch data.Type {
		case "doc", "feat":
			data.Group = "added"
		case "refactor":
			data.Group = "changed"
			data.Group = "fixed"
		case "chore", "ci", "build":
			data.Group = "other"
		}

		return data, true
	}
	return nil, false
}
