package webui

import (
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/core"
)

var currentPlanName string
var templatesPath string = "webui/templates"

type NodeMetaInfoUI struct {
	Path        string
	ShortPath   string
	LastRevSize int64
	AllRevSize  int64
	IsDir       bool
}

func Init(planName string) {
	if planName == "" {
		base.LogErr.Fatalln("Plan name should be defined to init web UI")
	}
	currentPlanName = planName

	http.HandleFunc("/static/", staticHandler)
	http.HandleFunc("/", mainHandler)
	http.ListenAndServe(":8080", nil)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) > 1 && parts[1] == "plan" {
		planName := ""
		if len(parts) > 2 {
			planName = parts[2]
		}

		if !core.IsBackupPlanExists(planName) {
			fmt.Fprintf(w, "Plan \"%s\" is not exists", planName)
			return
		}
		currentPlanName = planName

		plan, err := core.GetBackupPlan(planName)
		if err != nil {
			fmt.Fprintf(w, "Error occured with getting info about plan \"%s\": %v", planName, err)
			return
		}

		cmds := []string{"archived_list"}
		defaultCmd := cmds[0]

		cmd := ""
		if len(parts) > 3 {
			cmd = parts[3]
		}
		if cmd != "" {
			switch cmd {
			case "archived_list":
				cmd_ArchivedList(w, r, plan)
			default:
				http.NotFound(w, r)
			}
		} else {
			http.Redirect(w, r, "/plan/"+planName+"/"+defaultCmd, 302)
		}
	} else {
		http.Redirect(w, r, "/plan/"+currentPlanName, 302)
	}
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	data, err := base.Asset("webui" + r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(r.URL.Path))
	if mimeType != "" {
		w.Header().Add("Content-type", mimeType)
	}
	w.Write(data)
}

func cmd_ArchivedList(w http.ResponseWriter, r *http.Request, plan core.BackupPlan) {
	basePath := r.FormValue("basePath")
	archivedNodesMap := plan.GetArchivedNodesAllRevMap()
	workPathArchivedNodesMap := make(map[string]*NodeMetaInfoUI)
	for p, nodes := range archivedNodesMap {
		if basePath != "" && !base.IsPathInBasePath(basePath, p) {
			continue
		}

		flPath := base.GetFirstLevelPath(basePath, p)
		if flPath != "" {
			nodeUI, exists := workPathArchivedNodesMap[flPath]
			if !exists {
				nodeUI = &NodeMetaInfoUI{
					Path:  flPath,
					IsDir: nodes[0].IsDir() || flPath != p,
				}
				workPathArchivedNodesMap[flPath] = nodeUI
			}
			_, nodeUI.ShortPath = filepath.Split(nodeUI.Path)
			for i, node := range nodes {
				nodeUI.AllRevSize += node.Size()
				if i == len(nodes)-1 {
					nodeUI.LastRevSize += node.Size()
				}
			}
		}
	}
	workPathArchivedNodes := NodeMetaInfoUIList{}
	for _, nodeUI := range workPathArchivedNodesMap {
		workPathArchivedNodes = append(workPathArchivedNodes, nodeUI)
	}
	sort.Sort(workPathArchivedNodes)
	basePathList := []map[string]string{}

	basePathList = append(basePathList, map[string]string{
		"Path":      "",
		"ShortPath": "root",
	})
	if basePath != "" {
		parts := strings.Split(basePath, string(filepath.Separator))
		for i, part := range parts {
			if part == "" {
				continue
			}

			p := make(map[string]string)
			p["Path"] = strings.Join(parts[0:i+1], string(filepath.Separator))
			p["ShortPath"] = part
			basePathList = append(basePathList, p)
		}
	}

	tplData := struct {
		PlanName      string
		PathSeparator string
		NodesList     []*NodeMetaInfoUI
		BasePathList  []map[string]string
	}{}
	tplData.PlanName = plan.Name
	tplData.PathSeparator = string(filepath.Separator)
	tplData.NodesList = workPathArchivedNodes
	tplData.BasePathList = basePathList

	tplFuncMap := template.FuncMap{
		"filesizeHumanView": filesizeHumanView,
	}

	t, err := template.New("archived_list").Funcs(tplFuncMap).Parse(getTemplateSrc(templatesPath + "/archived_list.html"))
	if err != nil {
		fmt.Fprintf(w, "Error occured while parsing template: %v", err)
		return
	}

	err = t.Execute(w, tplData)
	if err != nil {
		fmt.Fprintf(w, "Error occured while parsing template: %v", err)
		return
	}
}

func getTemplateSrc(name string) string {
	data, err := base.Asset(name)
	if err != nil {
		base.LogErr.Fatalf("template file is not founded: %s\n", name)
	}
	return string(data)
}

func filesizeHumanView(size int64) string {
	var KB, MB, GB, TB float64
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024
	TB = GB * 1024

	sizef := float64(size)

	if sizef > TB {
		return strconv.FormatFloat(sizef/TB, 'f', 2, 64) + " T"
	} else if sizef > GB {
		return strconv.FormatFloat(sizef/GB, 'f', 2, 64) + " G"
	} else if sizef > MB {
		return strconv.FormatFloat(sizef/MB, 'f', 2, 64) + " M"
	} else if sizef > KB {
		return strconv.FormatFloat(sizef/KB, 'f', 2, 64) + " K"
	} else {
		return strconv.FormatInt(size, 10)
	}
}

type NodeMetaInfoUIList []*NodeMetaInfoUI

func (nl NodeMetaInfoUIList) Len() int {
	return len(nl)
}
func (nl NodeMetaInfoUIList) Swap(i, j int) {
	nl[i], nl[j] = nl[j], nl[i]
}
func (nl NodeMetaInfoUIList) Less(i, j int) bool {
	return (nl[i].IsDir && !nl[j].IsDir) ||
		(nl[i].IsDir == nl[j].IsDir && nl[i].ShortPath < nl[j].ShortPath)
}
