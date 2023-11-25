package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/andygrunwald/vdf"
)

var (
	steamPossibleFolderList = []string{"Program Files (x86)/Steam", "Program Files/Steam", "Steam"}
	steamInstalledFolder    string
	steamIconFolder         string
	installedGameIDList     []string
)

type GameInfo struct {
	Data map[string]AppIDInfo `json:"data"`
}

type AppIDInfo struct {
	Common struct {
		Clienticon    string `json:"clienticon"`
		Name          string `json:"name"`
		NameLocalized struct {
			Schinese string `json:"schinese"`
		} `json:"name_localized"`
		Type string `json:"type"`
	} `json:"common"`
}

type NeedData struct {
	icon  string
	name  string
	cname string
}

type LibraryVdf struct {
	Libraryfolders map[string]struct {
		Apps map[string]string `json:"apps"`
	} `json:"libraryfolders"`
}

func getDiskList() []string {
	diskList := []string{}
	for i := 67; i <= 90; i++ {
		disk := fmt.Sprintf("%c:/", i)
		if info, err := os.Stat(disk); err != nil {
			return diskList
		} else if info.IsDir() {
			diskList = append(diskList, disk)
		}
	}
	return diskList
}

func getGameInfo(appId string) NeedData {
	url := fmt.Sprintf("http://api.steamcmd.net/v1/info/%s", appId)
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != 200 {
		fmt.Print("请求失败，正在重试")
		getGameInfo(appId)
	}
	defer res.Body.Close()
	var gameInfo GameInfo
	body, _ := io.ReadAll(res.Body)
	json.Unmarshal(body, &gameInfo)
	if err != nil {
		fmt.Println(err)
	}
	var resData NeedData
	for _, appInfo := range gameInfo.Data {
		resData = NeedData{
			icon:  appInfo.Common.Clienticon,
			name:  appInfo.Common.Name,
			cname: appInfo.Common.NameLocalized.Schinese,
		}
	}
	return resData
}

func getIconFile(appId string, icon string) []byte {
	// http://cdn.cloudflare.steamstatic.com/
	iconReqUrl := fmt.Sprintf("http://cdn.akamai.steamstatic.com/steamcommunity/public/images/apps/%s/%s.ico", appId, icon)
	req, _ := http.NewRequest("GET", iconReqUrl, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != 200 {
		fmt.Print("下载失败，正在重试")
		getIconFile(appId, icon)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return body
}

func scanSteamInstallation() {
	if _, err := os.Stat(path.Join(steamInstalledFolder, "steam.exe")); err != nil {
		fmt.Print("自动检测Steam安装目录：")
	} else {
		fmt.Println("传入固定路径")
		steamIconFolder = path.Join(steamInstalledFolder, "steam/games")
		return
	}
	diskList := getDiskList()
	for _, disk := range diskList {
		for _, installFolder := range steamPossibleFolderList {
			if _, err := os.Stat(path.Join(disk, installFolder, "steam.exe")); err == nil {
				steamInstalledFolder = path.Join(disk, installFolder)
				steamIconFolder = path.Join(steamInstalledFolder, "steam/games")
				fmt.Printf("%s\n", steamInstalledFolder)
				return
			}
		}
	}
	fmt.Println("自动检测失败，请手动指定Steam安装目录：")
	for {
		steamFolder := ""
		if _, err := fmt.Scanf("%s\n", &steamFolder); err != nil {
			fmt.Printf("ERROR! %s", err)
		} else {
			if _, err := os.Stat(path.Join(steamFolder, "steam.exe")); err != nil {
				fmt.Println("输入的路径无效，请重新输入")
			} else {
				steamInstalledFolder = steamFolder
				steamIconFolder = path.Join(steamInstalledFolder, "steam/games")
				return
			}
		}
	}
}

func scanSteamGameId() {
	libraryVdfPath := path.Join(steamInstalledFolder, "steamapps/libraryfolders.vdf")
	if _, err := os.Stat(libraryVdfPath); err != nil {
		fmt.Printf("ERROR! [%s]读取失败", libraryVdfPath)
		return
	}
	libraryVdfFile, err := os.Open(libraryVdfPath)
	if err != nil {
		fmt.Printf("ERROR! [%s]读取失败", libraryVdfPath)
		return
	}
	defer libraryVdfFile.Close()
	if vdfMap, err := vdf.NewParser(libraryVdfFile).Parse(); err != nil {
		fmt.Printf("ERROR! 解析配置文件错误")
	} else {
		vdfJson, _ := json.Marshal(vdfMap)
		var vdfData LibraryVdf
		json.Unmarshal(vdfJson, &vdfData)
		for _, curLibrary := range vdfData.Libraryfolders {
			curLibIDList := []string{}
			for appKey := range curLibrary.Apps {
				curLibIDList = append(curLibIDList, fmt.Sprintf("%v", appKey))
			}
			installedGameIDList = append(installedGameIDList, curLibIDList...)
		}
	}
}

func reloadSteamIcon() {
	for _, appId := range installedGameIDList {
		info := getGameInfo(appId)
		displayName := info.name
		if info.cname != "" {
			displayName = info.cname
		}
		if info.icon != "" {
			if _, err := os.Stat(path.Join(steamIconFolder, info.icon+".ico")); err != nil {
				fmt.Printf("正在下载 %s 的图标文件... ", displayName)
				iconFile := getIconFile(appId, info.icon)
				file, err := os.Create(path.Join(steamIconFolder, info.icon+".ico"))
				if err != nil {
					fmt.Printf("创建文件错误\n")
				} else {
					defer file.Close()
					file.Write(iconFile)
					fmt.Printf("下载完成\n")
				}
			} else {
				fmt.Printf("存在 %s 的图标文件，跳过\n", displayName)
			}
		} else {
			fmt.Printf("%s 不存在图标\n", displayName)
		}
	}
	fmt.Print("修复完成")
}

func main() {
	if len(os.Args) > 1 {
		steamInstalledFolder = os.Args[1]
	}
	scanSteamInstallation()
	scanSteamGameId()
	reloadSteamIcon()
	fmt.Println(" 按任意键退出...")
	exit := make([]byte, 1)
	os.Stdin.Read(exit)
}
