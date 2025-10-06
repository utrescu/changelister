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
	Start plumbing.Hash
	Stop  plumbing.Hash
}


func ProcessTags(repo *git.Repository, tagTo string) ([]TagInfo, error) {
	listTags := GetRepoTags(repo)

	// Si no s'ha especificat etiqueta, les processem totes
	if tagTo == "" {
		return ProcessAllTags(repo, listTags)
	} 

	result := []TagInfo{}

	// L'etiqueta ha d'existir
	if !slices.Contains(listTags, tagTo) {
		log.Printf("L'etiqueta %s no existeix al repositori\n", tagTo)
		return nil, fmt.Errorf("L'etiqueta %s no existeix al repositori", tagTo)
	}

	tag1 := TagInfo{
		Name: tagTo,
		Start: GetNextTag(repo, listTags, tagTo),
		Stop: GetTagCommit(repo, tagTo),
	}
	result = append(result, tag1)
	return result, nil
}

func ProcessAllTags(repo *git.Repository, listTags []string) ([]TagInfo, error) {
	result := []TagInfo{}

	slices.Reverse(listTags)

	firstTag := TagInfo{
		Name:  "Unversioned",
		Start: GetNextTag(repo, listTags, ""),
		Stop:  GetLastCommit(repo),
	}
	result = append(result, firstTag)

	
	currentEnd := firstTag.Start

	for _, tag := range listTags {
		newTag := TagInfo{
			Name:  tag,
			Start: GetNextTag(repo, listTags, tag),
			Stop:  currentEnd,
		}
		currentEnd = newTag.Start
		result = append(result, newTag)
	}
	return result, nil
}

func GetNextTag(repo *git.Repository, listTags []string, tagTo string) plumbing.Hash {

	if tagTo == "" {
		// Si no s'ha especificat etiqueta, retornem el primer commit del repositori
		return GetTagCommit(repo, listTags[0])
	}

	for i := range listTags[:len(listTags)-1] {
		if listTags[i] == tagTo  {
			return GetTagCommit(repo, listTags[i+1])
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

func GetRepoTags(repo *git.Repository) []string {
	tags := []string{}

	tagrefs, err := repo.Tags()
	if err != nil {
		panic(err)
	}

	err = tagrefs.ForEach(func(t *plumbing.Reference) error {
		tags = append(tags, t.Name().Short())
		return nil
	})

	return tags
}
