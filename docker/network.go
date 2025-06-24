package docker

import (
    "context"
    "encoding/json"
    "net/http"
    "github.com/docker/docker/api/types/network"
    "github.com/docker/docker/client"
)

type CreateNetworkRequest struct {
    Name       string            `json:"name"`
    Driver     string            `json:"driver"`
    Internal   bool              `json:"internal"`
    Attachable bool              `json:"attachable"`
    EnableIPv6 bool              `json:"enable_ipv6"`
    Subnet     string            `json:"subnet"`
    Gateway    string            `json:"gateway"`
    Labels     map[string]string `json:"labels"`
    Options    map[string]string `json:"options"`
}

func CreateNetworkHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateNetworkRequest
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    ctx := context.Background()
    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cli.Close()

    ipamConfig := []network.IPAMConfig{}
    if req.Subnet != "" || req.Gateway != "" {
        cfg := network.IPAMConfig{}
        if req.Subnet != "" {
            cfg.Subnet = req.Subnet
        }
        if req.Gateway != "" {
            cfg.Gateway = req.Gateway
        }
        ipamConfig = append(ipamConfig, cfg)
    }

    options := network.CreateOptions{
        Driver:     req.Driver,
        Internal:   req.Internal,
        Attachable: req.Attachable,
        EnableIPv6: &req.EnableIPv6,
        Labels:     req.Labels,
        Options:    req.Options,
        IPAM: &network.IPAM{
            Driver: "default",
            Config: ipamConfig,
        },
    }

    resp, err := cli.NetworkCreate(ctx, req.Name, options)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"id": resp.ID})
}

func ListNetworksHandler(w http.ResponseWriter, r *http.Request) {
    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cli.Close()

    networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{"networks": networks})
}

type DeleteNetworkRequest struct {
    ID string `json:"id"`
}

func DeleteNetworkHandler(w http.ResponseWriter, r *http.Request) {
    var req DeleteNetworkRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
        http.Error(w, "Missing or invalid network id", http.StatusBadRequest)
        return
    }

    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cli.Close()

    err = cli.NetworkRemove(context.Background(), req.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Network deleted"})
}
