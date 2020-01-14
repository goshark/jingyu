package upload

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/glog"
)

func UploadMore(r *ghttp.Request,parent_path string,savename ...string) []string{

	SERVER_NAME := r.Server.GetName()
	if SERVER_NAME == ""{
		SERVER_NAME = "Public"
	}
	UPLOAD_ROOT := g.Config().GetString("upload-path")
	pic_paths := make([]string,0)
	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0{
		return pic_paths
	}
	//glog.Println(r.MultipartForm.File)
	//设置内存大小
	 if err := r.ParseMultipartForm(32 << 20);err != nil {
	 	glog.Error(err)
	 	return pic_paths
	 }
	//获取上传的文件组

	 filemap := r.MultipartForm.File
	 glog.Println(filemap)
	 for k,_ := range filemap{
		 files := r.MultipartForm.File[k]
		 length := len(files)

		 for i := 0; i < length; i++ {
			 //打开上传文件
			 file, err := files[i].Open()
			 defer file.Close()
			 if err != nil {
				 glog.Error(err)
			 }
			 buffer := make([]byte, files[i].Size)
			 file.Read(buffer)

			 //创建上传目录
			 if !gfile.Exists(UPLOAD_ROOT){
				 gfile.Mkdir(UPLOAD_ROOT)
			 }
			 pic_path := ""
			 if len(savename) > 0 {
				 pic_path =	"/"+SERVER_NAME+"/"+ parent_path +"/"+ savename[0]+gfile.Ext(gfile.Basename(files[i].Filename))
			 }else{
				 pic_path = "/"+SERVER_NAME+"/"+ parent_path +"/"+ gfile.Basename(files[i].Filename)
			 }

			 file_path := UPLOAD_ROOT+pic_path
			 //if gfile.Exists(file_path){
			 //	extfile,err := gfile.Open(file_path)
			 //	defer extfile.Close()
			 //	if err != nil {
			 //		glog.Error(err)
				//}
			 //	extfileinfo,err := extfile.Stat()
				// if err != nil {
				//	 glog.Error(err)
				// }
			 //
			 //	if extfileinfo.Size() != files[i].Size{
			 //
			 //		glog.Println("上传图片：",file_path)
				//	pic_paths = append(pic_paths, pic_path)
				//	gfile.PutBinContents(file_path, buffer)
				//}
			 //}else{
				 glog.Println("上传图片：",file_path)
				 pic_paths = append(pic_paths, pic_path)
			 	glog.Println("当前图片",i,pic_path)
				gfile.PutBytes(file_path,buffer)


			 //}

		 }
	 }
	 glog.Println("图片个数",len(pic_paths),pic_paths)
return pic_paths


}