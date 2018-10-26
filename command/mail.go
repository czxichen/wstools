package command

import (
	"github.com/czxichen/wstools/common/cli"
	"github.com/spf13/cobra"
)

// Mail 邮件命令
var Mail = &cobra.Command{
	Use: "mail",
	Example: `	发送邮件
	-u user -p passwd -H smtp.163.com:25 -f czxichen@163.com -t czxichen@163.com -c "Hello world"`,
	Short: "邮件发送",
	Long:  "使用smtp协议发送邮件,可以为文本格式或带附件发送",
	Run:   mailRun,
}

var mailConfig cli.MailConfig

func init() {
	Mail.PersistentFlags().StringVarP(&mailConfig.User, "user", "u", "", "指定登录的用户,不能为空")
	Mail.PersistentFlags().StringVarP(&mailConfig.Passwd, "passwd", "p", "", "指定用户密码,不能为空")
	Mail.PersistentFlags().StringVarP(&mailConfig.Host, "host", "H", "", "指定服务器地址端口")
	Mail.PersistentFlags().StringVarP(&mailConfig.Subject, "subject", "s", "", "指定邮件主题")
	Mail.PersistentFlags().StringVarP(&mailConfig.From, "from", "f", "", "指定发送者的地址,不能为空")
	Mail.PersistentFlags().StringVarP(&mailConfig.To, "to", "t", "", "指定接收者的地址,多地址使用','分割,不能为空")
	Mail.PersistentFlags().StringVarP(&mailConfig.Content, "content", "c", "", "指定邮件内容")
	Mail.PersistentFlags().StringVarP(&mailConfig.ContentPath, "Cpath", "C", "", "用文件内容做邮件内容,不能和-c同时使用")
	Mail.PersistentFlags().StringVarP(&mailConfig.Attachments, "attachments", "a", "", "指定附件路径,多个附件用','分割")
	Mail.PersistentFlags().StringVarP(&mailConfig.Type, "type", "T", "plain", "指定邮件格式:plain|html")

}

func mailRun(cmd *cobra.Command, args []string) {
	if err := cli.MailRun(&mailConfig); err != nil {
		cli.FatalOutput(1, "Send mail error:%s", err.Error())
	}
}
