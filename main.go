package main

import (
	"fmt"
	"github.com/robfig/cron"
	"os"
	"seeder/src/config"
	"seeder/src/datebase"
	"seeder/src/nexus"
	"seeder/src/qbittorrent"
	"strconv"
)

func main() {
	var db datebase.Client
	var nodes []nexus.Client
	var servers []qbittorrent.Server

	cron := cron.New()
	if cfg, err := config.GetConfig(); err == nil {
		db = datebase.NewClient(cfg.Db)
		for _, value := range cfg.Node {
			if value.Enable == true {
				node := nexus.NewClient(value.Source, value.Limit, value.Passkey, value.Rule)
				nodes = append(nodes, node)
			}
		}
		for _, value := range cfg.Server {
			if value.Enable == true {
				server := qbittorrent.NewClientWrapper(value.Endpoint, value.Username, value.Password, value.Remark, value.Rule)

				server.CalcEstimatedQuota()
				server.ServerClean(cfg, db)

				cron.AddFunc("@every 5s", func() { server.CalcEstimatedQuota() })
				cron.AddFunc("@every 1m", func() { server.AnnounceRace() })
				cron.AddFunc("@every 1m", func() { server.ServerClean(cfg, db) })
				cron.Start()

				servers = append(servers, server)
			}
		}
	} else {
		os.Exit(1)
	}

	for true {
		var ts []nexus.Torrent
		for _, node := range nodes {
			fmt.Println("start query")
			ts, _ = node.Get()
			for _, t := range ts {
				// 解决重复添加问题
				for _, server := range servers {
					server.CalcEstimatedQuota()
					if db.Get(t.GUID) == false {
						if Size, err := strconv.Atoi(t.Size); err == nil {
							if server.AddTorrentByURL(t.URL, Size, int(node.Rule.SpeedLimit * 1024 * 1024)) == true {
								fmt.Println("[" + server.Remark + "][添加]种子:" + t.Title)
								db.Insert(t.Title, t.GUID, t.URL)
							}
						}
					}
				}
				if db.Get(t.GUID) == false {
					//找遍了所有服务器(10次尝试),还是没法找到添加的,那么,就直接插入数据库不再尝试.
					fmt.Println("[忽略]种子:" + t.Title)
					db.Insert(t.Title, t.GUID, t.URL)
				}
			}
		}
	}
}
