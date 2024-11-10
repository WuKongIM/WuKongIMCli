build:
	docker build -t wukongimcli .
deploy:
	docker build -t wukongimcli .
	docker tag wukongimcli registry.cn-shanghai.aliyuncs.com/wukongim/wukongimcli:latest
	docker push registry.cn-shanghai.aliyuncs.com/wukongim/wukongimcli:latest	

