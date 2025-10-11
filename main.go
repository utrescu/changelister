package main

import (
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/utrescu/changelister/commits"
	"github.com/utrescu/changelister/config"
	"github.com/utrescu/changelister/tags"
	"github.com/utrescu/changelister/template"
)

func main() {

	configuration, err := config.LoadConfig()
	if err != nil {
		log.Printf("Error loading config: %s", err.Error())
		return
	}

	repo, err := git.PlainOpen(configuration.Path)
	if err != nil {
		log.Println(err)
		return
	}

	listTags, err := tags.ProcessTags(repo, configuration.Tag, configuration.DefaultTag)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	var changelogData []commits.ChangelogData

	for _, repoTags := range listTags {
		processedTag := commits.ProcessTagCommits(repo, repoTags, configuration.CommitTypes.Show)
		changelogData = append(changelogData, processedTag)
	}

	err = template.ProcessTemplate(configuration.Template.File, "changelog.md", changelogData)
	if err != nil {
		log.Printf("Error processing template: %s", err.Error())
		return
	}

}
