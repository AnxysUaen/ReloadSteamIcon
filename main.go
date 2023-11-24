package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/andygrunwald/vdf"
)

var (
	steamClientIconBaseURL  = "https://cdn.cloudflare.steamstatic.com/steamcommunity/public/images/apps/"
	steamPossibleFolderList = []string{"Program Files (x86)/Steam", "Program Files/Steam", "Steam"}
	steamInstalledFolder    string
	steamIconFolder         string
	installedGameIDList     []string
)

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

func getGameInfo(appId string) string {
	url := "http://api.steamcmd.net/v1/info/" + appId
	req, _ := http.NewRequest("GET", url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != 200 {
		return ""
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	fmt.Println(string(body))
	return string(body)
}

func scanSteamInstallation() {
	if _, err := os.Stat(path.Join(steamInstalledFolder, "steam.exe")); err != nil {
		fmt.Println("全盘扫描Steam安装目录：")
	}
	diskList := getDiskList()
	for _, disk := range diskList {
		for _, installFolder := range steamPossibleFolderList {
			if _, err := os.Stat(path.Join(disk, installFolder, "steam.exe")); err == nil {
				steamInstalledFolder = path.Join(disk, installFolder)
				steamIconFolder = path.Join(steamInstalledFolder, "steam/games")
				return
			}
		}
	}
	fmt.Println("未检测到Steam，请手动指定Steam安装目录：")
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
	libraryVdf := path.Join(steamInstalledFolder, "steamapps/libraryfolders.vdf")
	if _, err := os.Stat(libraryVdf); err != nil {
		fmt.Printf("ERROR! [%s]读取失败", libraryVdf)
		return
	} else {
		if file, err := os.Open(libraryVdf); err != nil {
			fmt.Printf("ERROR! [%s]读取失败", libraryVdf)
			return
		} else {
			defer file.Close()
			if vdfMapFile, err := vdf.NewParser(file).Parse(); err != nil {
				fmt.Printf("ERROR! 解析配置文件错误")
			} else {
				if vdfMap, ok := vdfMapFile["Console Sample v.1"].(map[string]interface{}); ok {
					if libMap, ok := vdfMap["libraryfolders"].(map[string]interface{}); ok {
						for _, curLibrary := range libMap {
							curLibIDList := []string{}
							if curLibraryMap, ok := curLibrary.(map[string]interface{}); ok {
								if appIdMap, ok := curLibraryMap["apps"].(map[string]interface{}); ok {
									for appKey, appValue := range appIdMap {
										fmt.Println(appKey)
										curLibIDList = append(curLibIDList, fmt.Sprintf("%v", appValue))
									}
									installedGameIDList = append(installedGameIDList, curLibIDList...)
								}
							}
						}
					}
				}
				fmt.Println(installedGameIDList)
				return
			}
		}
	}
}

func reloadSteamIcon() {
	for _, app_id := range installedGameIDList {
		// 请求接口获得信息
		getGameInfo(app_id)
		// 获取name用来显示
		// 获取clienticon下载图标
		// http://cdn.akamai.steamstatic.com/steamcommunity/public/images/apps/{{appid}}/{{icon}}.ico
		// 检查是否下载成功
		// 检查本地是否存在
		// 替换
		// 替换是否成功
		// 添加延迟以免被ban
	}
	return
}

func main() {
	scanSteamInstallation()
	scanSteamGameId()
	reloadSteamIcon()
}
