package worker

import (
	"log"
	"sdwan/common"
)

type Vnet struct {
}

func (vpn Vnet) Name() string {
	return "Vnet"
}

func SetupCommon(ep common.EndpointVO) (string, bool) {
	if ok := CreateInterface(ep); !ok {
		log.Println("Create interfaces failed.")
		return "failed: create interface", false
	}
	if ok := CreateNetns(ep.Vni); !ok {
		log.Println("Create netns failed.")
		return "failed: create netns", false
	}
	if ok := CreateVeth(ep); !ok {
		log.Println("Create veth failed.")
		return "failed: veth netns", false
	}
	if ok := CheckNetnsLink(ep.NetnsName(), ep.VethnName()); !ok {
		if ok := SetLinkNetns(ep.VethnName(), ep.NetnsName()); !ok {
			log.Printf("set vethn into netns Failed.")
			return "failed: set vethn netns", false
		}
		if ok := SetupVethNetns(ep); !ok {
			log.Println("Setup veth and netns failed.")
			return "failed: setup feth", false
		}
	}
	if ok := UpdateFDB(ep); !ok {
		log.Println("Update FDB failed.")
		return "failed: update fdb", false
	}
	return "success", true
}

func (vpn Vnet) SetVnetEndpoint(msg common.Message) {
	ep := common.EndpointVO{}
	if err := common.LoadBody(msg.Body, &ep); err == nil {
		if ep.Action == "ADD" {
			if errmsg, ok := SetupCommon(ep); !ok {
				Response(msg.ToResult(errmsg))
				return
			}
			common.WriteFile(ep.ConfPath(), ep.GenBgpContent())
			common.WriteFile(ep.RelfectorConfPath(), ep.GenBgpContentRrs())
			go StartBird(ep.NetnsName())
		} else {
			if ok := DeleteInterface(ep); !ok {
				log.Println("Delete interfaces error.")
				Response(msg.ToResult("failed: remove interface"))
				return
			}
			StopBird(ep.NetnsName())
		}
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
}

func (vpn Vnet) SetVnetEndpointReflector(msg common.Message) {
	ep := common.EndpointVO{}
	if err := common.LoadBody(msg.Body, &ep); err == nil {
		if ep.Action == "ADD" {
			if errmsg, ok := SetupCommon(ep); !ok {
				Response(msg.ToResult(errmsg))
				return
			}
			content := ep.GenRRBgpContent()
			common.WriteFile(ep.ConfPath(), content)
			go StartBird(ep.NetnsName())
		} else {
			if ok := DeleteInterface(ep); !ok {
				log.Println("Delete interfaces error.")
				Response(msg.ToResult("failed: remove interface"))
				return
			}
			StopBird(ep.NetnsName())
		}
		Response(msg.ToResult("success"))
		return
	}
	Response(msg.ToResult("failed: unknown error"))
}
