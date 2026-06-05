package internal

import (
    "xray-manual-svc/internal/app/config"
    "xray-manual-svc/internal/app/xrayconfig"
    "xray-manual-svc/internal/clients/pingproxy"
    "xray-manual-svc/internal/clients/xray"
)

func Bootstrap(cfg *config.Config) (*ProxyManager, error) {
    xrayCfg, err := xrayconfig.Load(cfg.Xray.ConfigPath)
    if err != nil {
        return nil, err
    }

    tags := xrayCfg.BalancerTags(cfg.Xray.BalancerTag)

    xrayClient, err := xray.New(cfg.Xray.XrayGRPCAddr, cfg.Xray.BalancerTag)
    if err != nil {
        return nil, err
    }

    pingClient, err := pingproxy.New(cfg.Xray.HTTPProxy)
    if err != nil {
        return nil, err
    }

    return NewProxyManager(tags, xrayClient, pingClient), nil
}
