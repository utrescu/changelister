package main

import (
	"log"

	"github.com/go-git/go-git/v5"
	commits "github.com/utrescu/changelister/commits"
	config "github.com/utrescu/changelister/config"
	tags "github.com/utrescu/changelister/tags"
	template "github.com/utrescu/changelister/template"
)

func main() {

	config, err := config.LoadConfig()
	if err != nil {
		log.Printf("Error loading config: %s", err.Error())
		return
	}

	// Obir el repositori
	repo, err := git.PlainOpen(config.Path)
	if err != nil {
		log.Println(err)
		return
	}

	listTags, err := tags.ProcessTags(repo, config.Tag, config.DefaultTag)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	changelogData := []commits.ChangelogData{}

	for _, tags := range listTags {

		// Processar les dades de l'etiqueta
		processedTag := commits.ProcessTagCommits(repo, tags, config.CommitTypes.Show)
		changelogData = append(changelogData, processedTag)

	}

	err = template.ProcessTemplate(config.Template.File, "changelog.md", changelogData)
	if err != nil {
		log.Printf("Error processing template: %s", err.Error())
		return
	}

}
