package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/n-boy/backuper/base"
	"github.com/n-boy/backuper/cmds"
	"github.com/n-boy/backuper/core"
)

func main() {
	base.InitApp(base.DefaultAppConfig)

	parseCmd()

	base.FinishApp()
}

func parseCmd() {
	var planName = flag.String("plan", "", "")
	var createPlan = flag.Bool("create-plan", false, "")
	cmd_flags := make(map[string]*bool)
	cmd_list := []string{"edit", "view", "status", "backup", "restore", "sync"}
	for _, cmd := range cmd_list {
		cmd_flags[cmd] = flag.Bool(cmd, false, "")
	}

	flag.Usage = func() {
		fmt.Printf("usage: %s --create-plan\n", os.Args[0])
		fmt.Printf("       %s --plan my_plan_name --<command>\n", os.Args[0])
		fmt.Println("possible commands:")
		for _, cmd := range cmd_list {
			fmt.Printf("    --%v\n", cmd)
		}
		// flag.PrintDefaults()
		os.Exit(0)
	}

	flag.Parse()

	if *createPlan {
		cmds.Create()
	} else if *planName != "" {
		plan, err := core.GetBackupPlan(*planName)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			cmd_selected := ""
			for cmd, sel := range cmd_flags {
				if *sel {
					if cmd_selected == "" {
						cmd_selected = cmd
					} else {
						fmt.Println("Only one command must be selected to run for plan\n")
						flag.Usage()
						return
					}
				}
			}
			switch cmd_selected {
			case "edit":
				cmds.Edit(plan)
			case "view":
				cmds.View(plan)
			case "status":
				cmds.Status(plan)
			case "backup":
				cmds.Backup(plan)
			case "restore":
				cmds.Restore(plan)
			case "sync":
				cmds.Sync(plan)
			case "":
				fmt.Println("One command must be selected to run for plan\n")
				flag.Usage()
				return
			}

		}
	} else {
		flag.Usage()
		return
	}

}

// command-line interface
// backuper --create-plan
// 	принимаем от пользователя все поля для создания нового плана, и создаем план
// backuper --plan PlanName <command>

// Процесс бекапа:
// - доливаем недокачанный архив
// - получаем список файлов под наблюдением
// - строим список архивированных файлов
// - вычисляем список файлов к архивации
// - для каждого чанка файлов:
// 	- формируем имя архива
// 	- создаем локально архив
// 	- создаем локально файл с метаданными
// 	- заливаем архив и метаданные в хранилище
// 	- перемещаем файл с метаданными в основную папку плана
// 	- удаляем локальный архив
// archive_1_20150405200819.zip

// archive_20150405_1.tsv
// file_name  modtime  size is_dir
