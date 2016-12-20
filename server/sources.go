package server

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/web"
	"sort"
)

type Sources struct {
	sourceMap map[string]Source
}

func (sources *Sources) load(client *client.Client) error {
	clientSources, err := client.ListSources()
	if err != nil {
		return err
	}

	sources.sourceMap = make(map[string]Source)

	for _, apiSource := range clientSources {
		source := Source{
			ID:   apiSource.ID,
			Name: apiSource.Name,
			Type: apiSource.SrcType.String(),

			Dimensions: Dimensions{
				Width:  apiSource.HSize,
				Height: apiSource.VSize,
			},
		}

		if source.Type == "input" {
			source.InputStatus = apiSource.InputVideoStatus.String()
		}

		sources.sourceMap[source.String()] = source
	}

	return nil
}

func (sources *Sources) Index(name string) (web.Resource, error) {
	if source, found := sources.sourceMap[name]; !found {
		return nil, nil
	} else {
		return source, nil
	}
}

func (sources *Sources) Get() (interface{}, error) {
	return sources.sourceMap, nil
}

type Source struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	InputStatus string     `json:"input_status,omitempty"`
	Dimensions  Dimensions `json:"dimensions"`
}

func (source Source) String() string {
	return fmt.Sprintf("%d", source.ID)
}

func (source Source) Get() (interface{}, error) {
	return source, nil
}

func (source Source) buildState(screens map[string]ScreenState) SourceState {
	sourceState := SourceState{Source: source}

	for screenName, screenState := range screens {
		for _, sourceName := range screenState.ProgramSources {
			if sourceName == source.String() {
				sourceState.ProgramScreens = append(sourceState.ProgramScreens, screenName)
			}
		}
		for _, sourceName := range screenState.PreviewSources {
			if sourceName == source.String() {
				sourceState.PreviewScreens = append(sourceState.PreviewScreens, screenName)
			}
		}
	}

	sort.Strings(sourceState.ProgramScreens)
	sort.Strings(sourceState.PreviewScreens)

	return sourceState
}

type SourceState struct {
	Source

	ProgramScreens []string `json:"program_screens,omitempty"`
	PreviewScreens []string `json:"preview_screens,omitempty"`
}

func (source SourceState) Get() (interface{}, error) {
	return source, nil
}
