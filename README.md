## 描述
gitlab的CI/CD一般是跑完test后触发images的构建，真正的部署会修改存放k8s yaml的项目，大部分都是修改images这一项的tag版本，然后提交代码被类似argocd这类集群管理服务钩子检测项目修改，自动完成部署。

## 功能
本服务就是为了完成跑完test并构建完镜像后，到修改yaml之间的步骤，也就是自动修改images的tag，为了完成这一步，需要在自动部署的项目加入一些shell脚本，让完成test和镜像构建的ci继续执行这个shell脚本，来调用服务来完成自动修改yaml的tag后提交代码

## 3个token
- ACCESS_TOKEN: 这个是gitlab提交的鉴权，可以直接使用token而不需要使用ssh，这个是通过gitlab变量${ACCESS_TOKEN}，在deploy/deploy.sh文件里执行shell时候传递，是在gitlab的后台设置的系统环境变量，这样加更加的安全
- AUTO_DEPLOY：这个是autodeploy服务的鉴权，在http/middleware/auth.go AUTH方法里可以看到，根据自己的需求更改。同样是通过设置gitlab的环境变量来实现
- RSA_PRIVATE_KEY：这个是ci-runner跑test和build需要的ssh鉴权，同样是提前设置在gitlab后台的系统变量，不过这个属于ci/cd的功能。

## 两个安全检查
- 上面的AUTO_DEPLOY的token加是基础的auth检查
- 同时我在WhitelistCheck()方法设置了ip白名单检查，服务只允许ci-runner机器来访问，这样就更加安全可靠。

## 使用
你需要了解gitlab-ci的知识，否则无法继续下
- 1.在了解完gitlab-ci后，可以查看.gitlab-ci.yml文件，这里是一个范例部署当前的autodeploy服务，在上线完成后同样要完成自举。
```shell
  script:
    - go build -o server # 编译
    - *build_script # 构建镜像
    - bash deploy/deploy.sh ${RESOURCES} # 触发部署
```
- 2.在完成编译和构建镜像后，开始执行自动部署的shell，我们来看看具体内容
```shell
http_code=$(curl -X POST -H "Content-Type:application/json " \
            -H "Authorization: Bearer ${AUTO_DEPLOY}" \
            --data-binary @$PAYLOAD \
            -s -o out.json \
            -w '%{http_code}' \
            http://autodeploy:80/api/v1/autodeploy/yaml/image/update) # 这里调用的就是自动部署的服务
```
可以看到最后请求的是http://autodeploy:80 域名，这个就是当前服务部署的域名，后面的path可以在项目的route里找到http/route/api.go
 - 3.resource.json是请求参数的组装，我在这里提供了一个范本，参数有gitlab配置的环境变量和gitlab-ci.yml里配置的环境变量
```shell
  variables:
    DOCKER_NS: algorithm-rls
    DEPLOY_REGISTRY: aliyuncs.com # 替换成自己的私有仓库
    ENV: dev
    CLUSTER_VERSION: default
    RESOURCES: resource.json
```
