package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sync"
	"time"

	. "autodeploy/util/object"
	utilYaml "autodeploy/util/yaml"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

type Meta struct {
	Author  string  `json:"author"`
	Message string  `json:"message"`
	Refer   string  `json:"refer"`
	App     string  `json:"app"`
	Team    string  `json:"team"`
	Token   string  `json:"token"`
	Items   []*Item `json:"items"`
}

type Item struct {
	Team      string `json:"team"`
	App       string `json:"app"`
	Component string `json:"component"`
	Env       string `json:"env"`
	Cluster   string `json:"cluster"`
	Kind      string `json:"kind"`
	Version   string `json:"version"`
	Image     string `json:"image"`
	Name      string `json:"name"` // deploy的name:deploy.Spec.Template.Spec.Containers下可能多个images,所有要匹配
	Yaml      string `json:"yaml"` // 直接指定的yaml文件，避免大家yaml文件命名不规范
}

var (
	once sync.Once
)

func init() {
	once.Do(func() {
	})
}

func UpdateYamlImages(c *gin.Context) {
	metrics := Metrics.Get(c)
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		InternalErr(c, "git body fail")
		return
	}
	var meta Meta
	if err := json.Unmarshal(data, &meta); err != nil {
		InternalErr(c, "json fail")
		return
	}
	log.Info().Msgf("meta: %v", meta)

	if !validate(meta) {
		InternalErr(c, "validate fail")
		return
	}

	err = gitoperate(meta)
	if err != nil {
		InternalErr(c, "git fail")
		return
	}

	// increment counter
	metrics.TotalReqsCounter.WithLabelValues(meta.Team, meta.App, meta.Author).Add(1)

	c.JSON(http.StatusOK, "success :)")
}

func validate(meta Meta) bool {
	for _, v := range meta.Items {
		if !inArray([]string{"dev", "prod"}, v.Env) {
			return false
		}
		if !regexpArguments(v.Yaml) {
			return false
		}
		if !regexpArguments(v.Env) {
			return false
		}
		if !regexpArguments(v.Image) {
			return false
		}
	}
	if !regexpArguments(meta.Refer) {
		return false
	}
	if !regexpArguments(meta.Author) {
		return false
	}
	if !regexpArguments(meta.Message) {
		return false
	}

	return true
}

func gitoperate(meta Meta) error {
	// repo url
	var gitRepoUrl string
	gitRepoUrl = meta.Team
	accesstoken := meta.Token

	wg, _ := errgroup.WithContext(context.Background())

	// Clone the Git repository
	r, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL: gitRepoUrl,
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth: &githttp.BasicAuth{
			Username: "fermi",
			Password: accesstoken,
		},
		Progress: os.Stdout,
	})
	if err != nil {
		log.Err(err).Msgf("git clone fail%v", err)
		return err
	}

	// Get the worktree
	worktree, err := r.Worktree()
	if err != nil {
		log.Err(err).Msgf("git worktree fail%v", err)
		return err
	}
	fmt.Println("worktree:", worktree)
	for index, item := range meta.Items {
		log.Info().Msgf("item - index: %v - %d", item, index)
		filename := item.Env + "/" + item.Yaml
		name := item.Name
		image := item.Image
		wg.Go(func() error {
			// Read the deployment.yaml file need 0755 to write
			file, err := worktree.Filesystem.OpenFile(filename, os.O_RDWR, 0755)
			if err != nil {
				log.Err(err).Msgf("git deployment fail%v", err)
				return err
			}

			defer func(file billy.File) {
				err := file.Close()
				if err != nil {

				}
			}(file)

			data, err := io.ReadAll(file)
			if err != nil {
				log.Err(err).Msgf("git ReadAll fail%v", err)
				return err
			}
			// Parse the deployment.yaml file
			var deploy map[any]any
			err = yaml.Unmarshal(data, &deploy)
			if err != nil {
				log.Err(err).Msgf("git yaml.Unmarshal fail%v", err)
				return err
			}

			log.Info().Msgf("Deployment - old: %v", deploy)
			// Modify the image field
			if spec, ok := deploy["spec"]; !ok {
				return xerrors.Errorf("deploy spec type fail:%v", deploy)
			} else if specMap, err := utilYaml.AnyToMap(spec); err != nil {
				return xerrors.Errorf("deploy specMap type fail:%v", deploy)
			} else if template, ok := specMap["template"]; !ok {
				return xerrors.Errorf("deploy template type fail:%v", deploy)
			} else if templateMap, err := utilYaml.AnyToMap(template); err != nil {
				return xerrors.Errorf("deploy templateMap type fail:%v", deploy)
			} else if specIn, ok := templateMap["spec"]; !ok {
				return xerrors.Errorf("deploy specIn type fail:%v", deploy)
			} else if specInMap, err := utilYaml.AnyToMap(specIn); err != nil {
				return xerrors.Errorf("deploy specInMap type fail:%v", deploy)
			} else if containers, ok := specInMap["containers"]; !ok {
				return xerrors.Errorf("deploy containers type fail:%v", deploy)
			} else if containersSlice, err := utilYaml.AnyToSlice(containers); err != nil {
				return xerrors.Errorf("deploy containersSlice type fail:%v", deploy)
			} else if len(containersSlice) > 0 {
				log.Info().Msgf("containersSlice:%v", containersSlice)
				for k, container := range containersSlice {
					if containerMap, err := utilYaml.AnyToMap(container); err != nil {
						return xerrors.Errorf("deploy single containerMap type fail:%v", deploy)
					} else {
						if containerMap["name"] == name {
							containerMap["image"] = image
							containersSlice[k] = containerMap
						}
					}
				}
				// todo Annotations will be cleared
				containers = containersSlice
				specInMap["containers"] = containers
				specIn = specInMap
				templateMap["spec"] = specIn
				template = templateMap
				specMap["template"] = template
				spec = specMap
				deploy["spec"] = spec
			} else {
				return xerrors.Errorf("deploy containersSlice is null:%v", deploy)
			}
			log.Info().Msgf("Deployment - new: %v", deploy)

			// Generate the new deployment.yaml file
			newData, err := yaml.Marshal(deploy)
			if err != nil {
				log.Err(err).Msgf("git yaml.Marshal fail%v", err)
				return err
			}

			// clean
			err = file.Truncate(0)
			if err != nil {
				log.Err(err).Msgf("git file Truncate fail%v", err)
				return err
			}
			_, err = file.Seek(0, 0)
			if err != nil {
				log.Err(err).Msgf("git file seek fail%v", err)
				return err
			}

			// new data
			_, err = file.Write(newData)
			if err != nil {
				log.Err(err).Msgf("file.Write() newData fail%v", err)
				return err
			}

			// Add the modification to the index
			_, err = worktree.Add(filename)
			if err != nil {
				log.Err(err).Msgf("git add fail%v", err)
				return err
			}

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		log.Err(err).Msgf("git routine fail%v", err)
		return err
	}

	// Commit the modification
	commit, err := worktree.Commit(fmt.Sprintf("%s:%s\nRefer:%s\nMessage:%s", meta.App, meta.Message, meta.Refer, meta.Message), &git.CommitOptions{
		Author: &object.Signature{
			Name:  "autodeploy",
			Email: meta.Author,
			When:  time.Now(),
		},
	})
	if err != nil {
		log.Err(err).Msgf("git commit fail%v", err)
		return err
	}
	// Push the commit to the remote repository
	err = r.Push(&git.PushOptions{
		Auth: &githttp.BasicAuth{
			Username: "test",
			Password: accesstoken,
		},
	})
	if err != nil {
		log.Err(err).Msgf("git Push fail%v", err)
		return err
	}
	// Print the commit information
	fmt.Println("Commit:", commit.String())

	return nil
}

func InternalErr(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": msg})
}

func inArray[T any](t []T, element T) bool {
	for _, v := range t {
		if reflect.DeepEqual(v, element) {
			return true
		}
	}
	return false
}

// regexpArguments
// @param str string "<example>"
func regexpArguments(str string) bool {
	pattern := "^<.*>$"
	reg, err := regexp.Compile(pattern)
	if err != nil {
		log.Err(err).Msgf("regexp fail%v", err)
		return false
	}
	// "<example>" is placeholder，match is false
	if reg.MatchString(str) {
		return false
	}
	return true
}
