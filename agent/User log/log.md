# 问题2.11

1. 当页面超过一整页在最后书写的时候，会跳转到最开头 [x]

2. 双界面预览界面应该同步，我经常在markdown界面写作但是预览界面还在前面 [x]

3. 文件夹导航栏应该可以收起 [x]

4. 文件夹应该可以持久化保存，而不是每次一启动就是空界面 [x]



# 问题 2.12

- 在用户设置api后没有长效保存的措施，导致每次启动软件时候都需要重新输api
- 是否应该添加一个按钮测试OpenAI API接口是否能有响应
- 我不知道这个ai是否真的在工作，我打开文件后提供了api链接和密钥但是我没有任何反馈并且api网站也没有接收到api响应


# 问题 2.13
- chat界面的api配置和供应商配置缺失,最好可以做多个供应商持久化保存
- RAG embedding模型选择没有正确持久化，模型选择似乎只会用text-embedding-3-small
```
ERR | Failed to process document 编程语言/C++/基础语法/程序结构流程.md: embedding generation failed: OpenAI error: The model `text-embedding-3-small` does not exist or you do not have access to it. (type: invalid_request_error, code: model_not_found)
```
- 对于embedding模型的测试可以多传入一些字符，较少字符可能供应商不会答复
- 力导向图库完全不加载不可用
- 相关笔记侧面图完全不可用



