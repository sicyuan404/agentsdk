# 部署指南

## GitHub Pages 部署

### 自动部署

文档已配置GitHub Actions自动部署：

1. **触发条件**：
   - 推送到`main`分支
   - 修改`docs/**`目录下的文件

2. **部署流程**：
   ```
   推送代码 → GitHub Actions → 构建 → 部署到gh-pages
   ```

3. **访问地址**：
   https://wordflowlab.github.io/agentsdk/

### 首次配置

#### 步骤1：启用GitHub Pages

1. 进入仓库：https://github.com/wordflowlab/agentsdk
2. Settings → Pages
3. Source选择"GitHub Actions"
4. 保存

#### 步骤2：推送代码

```bash
cd /Users/coso/Documents/dev/ai/wordflowlab/agentsdk

# 添加文档文件
git add docs/
git add .github/workflows/deploy-docs.yml

# 提交
git commit -m "docs: 添加完整文档站点

- Nuxt 3 + Nuxt Content文档框架
- 包含介绍、核心概念、实战指南、API参考
- 配置GitHub Actions自动部署
- 支持中文、代码高亮、深色模式
"

# 推送到main分支
git push origin main
```

#### 步骤3：查看部署

1. 访问Actions标签：https://github.com/wordflowlab/agentsdk/actions
2. 查看"Deploy to GitHub Pages" workflow
3. 等待构建完成（约2-3分钟）
4. 访问：https://wordflowlab.github.io/agentsdk/

## 本地测试部署

在推送前，建议先本地测试：

```bash
cd docs

# 生成静态文件
npm run generate

# 预览
npm run preview
```

访问：http://localhost:3000/agentsdk/

## 更新文档

修改文档后：

```bash
# 修改docs/content/下的md文件
vim docs/content/1.introduction/1.overview.md

# 本地预览
cd docs && npm run dev

# 提交并推送
git add docs/
git commit -m "docs: 更新xxx文档"
git push
```

自动部署会在几分钟内完成。

## 故障排查

### 问题1：Actions失败

检查：
1. `.github/workflows/deploy-docs.yml`文件是否正确
2. `docs/package.json`依赖是否完整
3. 查看Actions日志获取详细错误

### 问题2：页面404

检查：
1. GitHub Pages是否启用
2. Source是否选择"GitHub Actions"
3. baseURL是否配置为`/agentsdk/`

### 问题3：样式丢失

检查：
1. `nuxt.config.ts`中`app.baseURL`配置
2. 静态资源路径是否正确
3. 图片使用`/agentsdk/images/xxx`格式

## 自定义域名

如需使用自定义域名：

1. 在`docs/public/`下创建`CNAME`文件
2. 写入域名（如：docs.example.com）
3. 配置DNS记录指向GitHub Pages
4. 推送代码

## 缓存清理

如果更新后看不到变化：

1. 清除浏览器缓存（Ctrl+Shift+R）
2. 等待CDN刷新（可能需要几分钟）
3. 检查是否真的推送成功

## 监控

查看部署状态：
- Actions: https://github.com/wordflowlab/agentsdk/actions
- Pages设置: https://github.com/wordflowlab/agentsdk/settings/pages

## 支持

遇到问题：
1. 查看[GitHub Actions文档](https://docs.github.com/en/actions)
2. 查看[GitHub Pages文档](https://docs.github.com/en/pages)
3. 在仓库提Issue
