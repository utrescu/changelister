package tags

import (
	"fmt"
	"log"
	"slices"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type TagInfo struct {
	Name string
	Message string
	Date string
	Start plumbing.Hash
	Stop  plumbing.Hash
}

func getTag(llista []TagInfo, name string) (TagInfo, error) {
	for _, tag := range llista {
		if tag.Name == name {
			return tag, nil
		}
	}
	return TagInfo{}, fmt.Errorf("L'etiqueta %s no existeix", name)
}


func ProcessTags(repo *git.Repository, tagTo string, defaultTag string) ([]TagInfo, error) {
	listTags := GetRepoTags(repo)

	// Si no s'ha especificat etiqueta, les processem totes
	if tagTo == "" {
		return ProcessAllTags(repo, listTags, defaultTag)
	} 

	result := []TagInfo{}

	// L'etiqueta ha d'existir
	currentTag, err := getTag(listTags,tagTo)
	if err != nil {
		log.Printf("L'etiqueta %s no existeix al repositori\n", tagTo)
		return nil, fmt.Errorf("L'etiqueta %s no existeix al repositori", tagTo)
	}

	tag1 := TagInfo{
		Name: currentTag.Name,
		Date: currentTag.Date,
		Message: currentTag.Message,
		Start: GetNextTag(repo, listTags, tagTo),
		Stop:  GetTagCommit(repo, tagTo),
	}
	result = append(result, tag1)
	return result, nil
}

func ProcessAllTags(repo *git.Repository, listTags []TagInfo, defaultTag string) ([]TagInfo, error) {
	result := []TagInfo{}

	slices.Reverse(listTags)

	firstTag := TagInfo{
		Name:  defaultTag,
		Start: GetNextTag(repo, listTags, ""),
		Stop:  GetLastCommit(repo),
	}
	result = append(result, firstTag)

	
	currentEnd := firstTag.Start

	for _, tag := range listTags {
		newTag := TagInfo{
			Name:  tag.Name,
			Message: tag.Message,
			Date: tag.Date,
			Start: GetNextTag(repo, listTags, tag.Name),
			Stop:  currentEnd,
		}
		currentEnd = newTag.Start
		result = append(result, newTag)
	}
	return result, nil
}

func GetNextTag(repo *git.Repository, listTags []TagInfo, tagTo string) plumbing.Hash {

	if tagTo == "" {

		if (len(listTags) == 0) {
			// Si no hi ha etiquetes, retornem el primer commit del repositori
			return getFirstCommit(repo)
		} 
		// Si no s'ha especificat etiqueta, retornem el primer commit del repositori
		return GetTagCommit(repo, listTags[0].Name)
	}

	for i := range listTags[:len(listTags)-1] {
		if listTags[i].Name == tagTo {
			return GetTagCommit(repo, listTags[i+1].Name)
		}
	}
	// Si no n'hi ha retornem el primer commit del repostori
	return getFirstCommit(repo)
}

func getFirstCommit(repo *git.Repository) plumbing.Hash {

	// Obté la referència de HEAD
	ref, err := repo.Head()
	if err != nil {
		log.Fatalf("Error obtenint HEAD: %v", err)
	}

	// Obté l’objecte commit al qual apunta HEAD
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Fatalf("Error obtenint commit: %v", err)
	}

	// Itera fins arribar al primer commit
	for commit.NumParents() > 0 {
		parent, err := commit.Parent(0)
		if err != nil {
			log.Fatalf("Error obtenint parent: %v", err)
		}
		commit = parent
	}

	return commit.Hash
}


func GetTagCommit(repo *git.Repository, tag string) plumbing.Hash {
	// Obté la referència de l’etiqueta
	ref, err := repo.Tag(tag)
	if err != nil {
		log.Fatalf("Error obtenint etiqueta %s: %v", tag, err)
	}

	// Obté l’objecte tag al qual apunta la referència
	tagObj, err := repo.TagObject(ref.Hash())
	if err != nil {
		// Si no és un objecte tag, pot ser un commit directe
		hash := ref.Hash()
		return hash
	}

	// Obté l’objecte commit al qual apunta l’objecte tag
	commit, err := tagObj.Commit()
	if err != nil {
		log.Fatalf("Error obtenint commit de l’etiqueta %s: %v", tag, err)
	}
	return commit.Hash
}

func GetLastCommit(repo *git.Repository) plumbing.Hash {
	ref, err := repo.Head()
	if err != nil {
		log.Fatalf("Error obtenint HEAD: %v", err)
	}

	// Obté l’objecte commit al qual apunta HEAD
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Fatalf("Error obtenint commit: %v", err)
	}
	return commit.Hash
}

func GetRepoTags(repo *git.Repository) []TagInfo {
	tags := []TagInfo{}

	tagrefs, err := repo.Tags()
	if err != nil {
		panic(err)
	}

	err = tagrefs.ForEach(func(t *plumbing.Reference) error {

		theTag, err := repo.TagObject(t.Hash())
		if err != nil {
			log.Printf("Error obtenint etiqueta %s: %v", t.Name().Short(), err)
			return nil
		}

		tags = append(tags, 
			TagInfo {
			Name: t.Name().Short(),
			Message: theTag.Message,
			Date: theTag.Tagger.When.String(),
		})
		return nil
	})

	return tags
}
