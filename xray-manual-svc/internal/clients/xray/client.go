package xray

import (
    "context"

    routerCommand "github.com/xtls/xray-core/app/router/command"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type BalancerStatus struct {
    Override        string
    PrincipleTarget string
}

type Client struct {
    balancerTag string
    conn        *grpc.ClientConn
    router      routerCommand.RoutingServiceClient
}

func New(addr, balancerTag string) (*Client, error) {
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }

    return &Client{
        balancerTag: balancerTag,
        conn:        conn,
        router:      routerCommand.NewRoutingServiceClient(conn),
    }, nil
}

func (c *Client) Close() error {
    return c.conn.Close()
}

func (c *Client) GetStatus(ctx context.Context) (*BalancerStatus, error) {
    resp, err := c.router.GetBalancerInfo(ctx, &routerCommand.GetBalancerInfoRequest{
        Tag: c.balancerTag,
    })
    if err != nil {
        return nil, err
    }

    status := &BalancerStatus{}
    if resp.Balancer.Override != nil {
        status.Override = resp.Balancer.Override.Target
    }
    if resp.Balancer.PrincipleTarget != nil && len(resp.Balancer.PrincipleTarget.Tag) > 0 {
        status.PrincipleTarget = resp.Balancer.PrincipleTarget.Tag[0]
    }

    return status, nil
}

func (c *Client) SetTarget(ctx context.Context, tag string) error {
    _, err := c.router.OverrideBalancerTarget(ctx, &routerCommand.OverrideBalancerTargetRequest{
        BalancerTag: c.balancerTag,
        Target:      tag,
    })
    return err
}

func (c *Client) ResetTarget(ctx context.Context) error {
    return c.SetTarget(ctx, "")
}
