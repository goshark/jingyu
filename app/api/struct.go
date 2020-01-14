package api

import (
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"strings"
)

type Apis struct {}

/*
	数据库实例
*/
func (api *Apis) db() gdb.DB{
	db := g.DB()
	//开启调试模式,prod环境下请关闭。默认为关闭
	db.SetDebug(true)
	return db
}

/*
	删除服务器上的图片
*/
func RemovePic(pic_paths string){

	if len(pic_paths) > 0 && pic_paths != ""{
		pic_bmp := []string{"bmp","jpg","png","tif","gif","pcx","tga","exif","fpx","svg","psd","cdr","pcd","dxf","ufo","eps","ai","raw","WMF","webp"}
		upload_path := g.Config().GetString("upload-path")
		pic_path_arr := strings.Split(pic_paths,";")

		for _,pic := range pic_path_arr{
			//pic为数据库逻辑路径
			//file_path 物理路径
			file_path := upload_path+pic
			//获取文件扩展
			file_ext := gfile.Ext(file_path)
			if file_ext == ""{//文件不存在
				glog.Error("删除文件错误：文件不存在!,",file_path)
				return
			}
			//检测该文件是否是图片
			for _,v := range pic_bmp{
				if file_ext == "."+v{//匹配格式
					err := gfile.Remove(file_path)
					if err != nil {
						glog.Error("删除文件错误：",err)
						return
					}
				}
			}
		}
	}
}

/*
	根据身份证号计算年龄
*/
func GetUserAgeByIdcard(idcard string)int{
	//获取当前年份
	current_year := gtime.Now().Year()
	//截取身份证年份
	born_year := gconv.Int(gstr.SubStr(idcard,6,4))
	age := current_year - born_year
	return age
}
