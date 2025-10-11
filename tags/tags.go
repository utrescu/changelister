package tags

import (
	"fmt"
	"log"
	"slices"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

type TagInfo struct {
	Name    string
	Message string
	Date    string
	Start   plumbing.Hash
	Stop    plumbing.Hash
}

func getTag(taglist []TagInfo, name string) (TagInfo, error) {
	for _, tag := range taglist {
		if tag.Name == name {
			return tag, nil
		}
	}
	return TagInfo{}, fmt.Errorf("tag %s don't exists", name)
}

func ProcessTags(repo *git.Repository, tagTo string, defaultTag string) ([]TagInfo, error) {
	listTags := GetRepoTags(repo)

	if tagTo == "" {
		return ProcessAllTags(repo, listTags, defaultTag)
	}

	var result []TagInfo
	currentTag, err := getTag(listTags, tagTo)
	if err != nil {
		log.Printf("Inexistent tag %s in repository\n", tagTo)
		return nil, fmt.Errorf("inexistent tag %s in repository", tagTo)
	}

	tag1 := TagInfo{
		Name:    currentTag.Name,
		Date:    currentTag.Date,
		Message: currentTag.Message,
		Start:   GetNextTag(repo, listTags, tagTo),
		Stop:    GetTagCommit(repo, tagTo),
	}
	result = append(result, tag1)
	return result, nil
}

func ProcessAllTags(repo *git.Repository, listTags []TagInfo, defaultTag string) ([]TagInfo, error) {
	var result []TagInfo

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
			Name:    tag.Name,
			Message: tag.Message,
			Date:    tag.Date,
			Start:   GetNextTag(repo, listTags, tag.Name),
			Stop:    currentEnd,
		}
		currentEnd = newTag.Start
		result = append(result, newTag)
	}
	return result, nil
}

func GetNextTag(repo *git.Repository, listTags []TagInfo, tagTo string) plumbing.Hash {

	if tagTo == "" {

		if len(listTags) == 0 {
			// In the repo don't have tags, return the first commit
			return getFirstCommit(repo)
		}
		// If no tags are found, return the last commit
		return GetTagCommit(repo, listTags[0].Name)
	}

	for i := range listTags[:len(listTags)-1] {
		if listTags[i].Name == tagTo {
			return GetTagCommit(repo, listTags[i+1].Name)
		}
	}
	// If it's the last tag
	return getFirstCommit(repo)
}

func getFirstCommit(repo *git.Repository) plumbing.Hash {

	ref, err := repo.Head()
	if err != nil {
		log.Fatalf("Error obtaining HEAD: %v", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Fatalf("Commit not found: %v", err)
	}

	for commit.NumParents() > 0 {
		parent, err := commit.Parent(0)
		if err != nil {
			log.Fatalf("parent not found: %v", err)
		}
		commit = parent
	}

	return commit.Hash
}

func GetTagCommit(repo *git.Repository, tag string) plumbing.Hash {
	ref, err := repo.Tag(tag)
	if err != nil {
		log.Fatalf("Label not found %s: %v", tag, err)
	}

	tagObj, err := repo.TagObject(ref.Hash())
	if err != nil {
		hash := ref.Hash()
		return hash
	}

	commit, err := tagObj.Commit()
	if err != nil {
		log.Fatalf("Error obtenint commit de l’etiqueta %s: %v", tag, err)
	}
	return commit.Hash
}

func GetLastCommit(repo *git.Repository) plumbing.Hash {
	ref, err := repo.Head()
	if err != nil {
		log.Fatalf("Error obtaining HEAD: %v", err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Fatalf("commit not found: %v", err)
	}
	return commit.Hash
}

func GetRepoTags(repo *git.Repository) []TagInfo {
	var tags []TagInfo

	tagrefs, err := repo.Tags()
	if err != nil {
		panic(err)
	}

	err = tagrefs.ForEach(func(t *plumbing.Reference) error {

		theTag, err := repo.TagObject(t.Hash())
		if err != nil {
			log.Printf("Label not found %s: %v", t.Name().Short(), err)
			return nil
		}

		tags = append(tags,
			TagInfo{
				Name:    t.Name().Short(),
				Message: theTag.Message,
				Date:    theTag.Tagger.When.String(),
			})
		return nil
	})

	return tags
}
