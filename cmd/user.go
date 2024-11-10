package cmd

import (
	"fmt"
	"log"

	wkproto "github.com/WuKongIM/WuKongIMGoProto"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

type userVar struct {
	num    int    // 数量
	prefix string // 用户前缀
}

type userCMD struct {
	ctx     *WuKongIMContext
	api     *API
	userVar *userVar
}

func newUserCMD(ctx *WuKongIMContext) *userCMD {
	u := &userCMD{
		ctx:     ctx,
		api:     NewAPI(),
		userVar: &userVar{},
	}
	return u
}

func (u *userCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "user",
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "create user",
		RunE:  u.runCreate,
	}

	cmd.AddCommand(create)

	u.initCreateVar(create)

	return cmd
}
func (u *userCMD) initCreateVar(cmd *cobra.Command) {
	cmd.Flags().StringVar(&u.userVar.prefix, "prefix", "usr", "用户前缀")
	cmd.Flags().IntVar(&u.userVar.num, "num", 0, "用户数量")
}

func (u *userCMD) runCreate(cmd *cobra.Command, args []string) error {
	u.api.SetBaseURL(u.ctx.opts.ServerAddr)

	uiprogress.Start()
	defer uiprogress.Stop()
	var progress *uiprogress.Bar
	progressNum := u.userVar.num
	if u.userVar.num <= 0 {
		progressNum = 1
	}
	progress = uiprogress.AddBar(progressNum).AppendCompleted().PrependElapsed()
	progress.Width = progressWidth()
	log.Printf("Starting create user, num: %d", u.userVar.num)

	state := "Setup     "
	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	// 创建用户
	for i := 0; i < u.userVar.num; i++ {
		uid := fmt.Sprintf("%s%d", u.userVar.prefix, i)
		err := u.api.UpdateToken(uid, "test", wkproto.APP, 0)
		if err != nil {
			return err
		}
		progress.Incr()
	}
	state = "Finished  "

	return nil
}
