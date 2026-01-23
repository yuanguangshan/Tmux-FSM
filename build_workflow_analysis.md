# Yuangs VSCode Extension 构建流程详细分析

## 完整构建流程概览

构建脚本 `compile.sh` 是一个自动化构建和打包工具，用于将 TypeScript/AssemblyScript 源代码编译为生产就绪的 VS Code 扩展包。

---

## 阶段 1: 环境准备与检查

### 步骤 1.1: 查找 Node.js 和 npm

**作用：**
- 自动检测系统中的 Node.js 和 npm 安装位置
- 支持多种安装方式（Homebrew、NVM、Volta、fnm 等）
- 设置正确的 PATH 环境变量

**实现方式：**
```bash
# 检查常见路径：
/usr/local/bin
/opt/homebrew/bin
~/.nvm/versions/node/*/bin
~/.volta/bin
~/.fnm/node-versions/*/installation/bin
```

**输出示例：**
```
🔍 正在查找 Node.js 和 npm...
✅ 找到 Node.js: /usr/local/bin/node
✅ 找到 npm: /usr/local/bin/npm
```

---

### 步骤 1.2: 显示版本信息

**作用：**
- 验证 Node.js 版本兼容性
- 显示 npm 版本用于调试
- 确保环境符合项目要求

**输出示例：**
```
📦 Node.js 版本:
v22.6.0

📦 npm 版本:
10.8.2
```

---

### 步骤 1.3: 检查 vsce 打包工具

**作用：**
- 验证 VS Code Extension (vsce) 打包工具是否已安装
- 如果未安装，自动安装 `@vscode/vsce`
- vsce 是创建 .vsix 扩展包的官方工具

**安装命令：**
```bash
npm install -g @vscode/vsce
```

**输出示例：**
```
🔍 检查 vsce (VSCE 打包工具)...
✅ vsce 已安装: /Users/ygs/.nvm/versions/node/v22.14.0/bin/vsce
```

---

## 阶段 2: 代码编译与构建

### 步骤 2.1: 编译 AssemblyScript 代码

**作用：**
- 将 AssemblyScript 源代码编译为 WebAssembly (WASM) 模块
- 生成 debug 和 release 两个版本
- WASM 用于实现高性能的沙箱执行环境

**执行命令：**
```bash
npm run asbuild
# 实际执行：
# - npm run asbuild:debug
# - npm run asbuild:release
```

**编译目标：**
```
src/engine/agent/governance/sandbox/core.as.ts
```

**生成文件：**
```
build/release.wasm (4.83 KB)
build/debug.wasm
```

**为什么需要 AssemblyScript？**
- TypeScript 语法的 WASM 编译器
- 提供接近原生的性能
- 实现安全的代码隔离和沙箱环境
- 用于 AI Agent 的执行控制和治理逻辑

**输出示例：**
```
🏗️  步骤 2: 执行完整构建流程
   ├── 子步骤 2.1: 编译 AssemblyScript 代码...
         - 编译 src/engine/agent/governance/sandbox/core.as.ts 为 debug 和 release 版本
         - 生成 WebAssembly 模块供沙箱环境使用

> yuangs-vscode@1.0.5 asbuild
> npm run asbuild:debug && npm run asbuild:release

> yuangs-vscode@1.0.5 asbuild:debug
> asc src/engine/agent/governance/sandbox/core.as.ts --target debug

> yuangs-vscode@1.0.5 asbuild:release
> asc src/engine/agent/governance/sandbox/core.as.ts --target release

         ✅ AssemblyScript 编译完成
```

---

### 步骤 2.2: 捆绑和优化代码

**作用：**
- 使用 Webpack 将所有 JavaScript/TypeScript 模块打包成单个文件
- 执行代码优化和压缩（production 模式）
- 复制 webview 资源文件到输出目录

**执行命令：**
```bash
npm run bundle
# 实际执行：
# webpack --mode production
# mkdir -p dist/webview
# cp src/vscode/webview/sidebar.html dist/webview/
# cp node_modules/marked/marked.min.js dist/webview/
```

**Webpack 打包详情：**

**生成的文件：**
```
dist/vscode/extension.js (451 KB) - 主扩展入口
dist/webview/sidebar.html - 侧边栏界面
dist/webview/marked.min.js - Markdown 渲染库
```

**打包内容分析：**
```
asset extension.js 451 KiB [compared for emit] [minimized]

模块统计：
- Node modules: 710 KiB (108 modules)
- 源代码 (src/): 188 KiB
  - engine/agent/: 135 KiB (18 modules)
  - vscode/: 36.8 KiB
    - ChatViewProvider.ts: 17 KiB
    - extension.ts: 2.64 KiB
    - askAI.ts: 3.97 KiB
  - engine/ai/client.ts: 4.98 KiB
  - runtime/vscode/VSCodeExecutor.ts: 6.77 KiB
```

**为什么需要打包？**
- 减少文件数量，提高加载速度
- 代码压缩减小体积
- 模块依赖解析和优化
- Tree-shaking 移除未使用的代码

**输出示例：**
```
   ├── 子步骤 2.2: 捆绑和优化代码...
         - 使用 Webpack 将所有模块捆绑成单个 extension.js 文件
         - 复制 webview 资源文件 (HTML, JS) 到 dist/webview/ 目录

> yuangs-vscode@1.0.5 bundle
> webpack --mode production && mkdir -p dist/webview && ...

asset extension.js 451 KiB [compared for emit] [minimized]
webpack 5.104.1 compiled successfully in 12770 ms

         ✅ 代码捆绑完成
```

---

## 阶段 3: 扩展打包

### 步骤 3.1: 准备打包环境

**作用：**
- 验证 package.json 的完整性
- 检查扩展清单（extension manifest）
- 确保所有必需的资源文件存在
- 执行 prepublish 脚本重新构建

**vsce 预检查项：**
- package.json 必需字段（name, version, publisher, engines 等）
- 扩展图标（如果有）
- 许可证文件
- README 文件

**输出示例：**
```
📦 步骤 3: 执行打包流程
   ├── 子步骤 3.1: 准备打包环境...
         - 验证 package.json 中的必要字段
         - 检查扩展清单文件
         - 确保所有必需的资源文件存在
         ✅ 打包环境准备就绪
```

---

### 步骤 3.2: 执行 vsce 打包命令

**作用：**
- 创建 VSIX (Visual Studio Extension) 安装包
- 收集所有扩展文件（源码、编译产物、资源等）
- 生成扩展清单和元数据
- 压缩为 .vsix 文件用于分发

**执行命令：**
```bash
npm run package
# 实际执行：
# vsce package
```

**打包流程细节：**

1. **重新执行 prepublish 脚本**
   ```
   Executing prepublish script 'npm run vscode:prepublish'...
   ```
   - 确保打包前代码是最新的
   - 再次运行完整的编译和打包流程

2. **收集文件**
   ```
   INFO  Files included in the VSIX:
   ```
   - 源代码文件
   - 编译产物
   - 配置文件
   - 文档
   - 资源文件

3. **生成 VSIX 包**
   ```
   DONE  Packaged: /Users/ygs/yuangs-vscode/yuangs-vscode-1.0.5.vsix
   ```

**打包内容清单：**
```
yuangs-vscode-1.0.5.vsix (23 files, 431.62 KB)
├── [Content_Types].xml
├── extension.vsixmanifest
└── extension/
    ├── package.json (3.33 KB)
    ├── build/release.wasm (4.83 KB)
    ├── dist/vscode/extension.js (451.45 KB)
    ├── dist/webview/sidebar.html (65.2 KB)
    ├── dist/webview/marked.min.js (48.55 KB)
    ├── .ai/context.json (147.52 KB)
    ├── README.md (4.12 KB)
    ├── LICENSE.txt (1.04 KB)
    ├── policy.yaml (0.56 KB)
    └── ... (其他文档和配置文件)
```

**vsce 打包过程：**
1. 读取 package.json 获取扩展信息
2. 根据 files 字段或 .vscodeignore 确定包含文件
3. 创建扩展清单（extension.vsixmanifest）
4. 将所有文件打包为 ZIP 格式（.vsix）
5. 验证包的完整性

**输出示例：**
```
   ├── 子步骤 3.2: 执行 vsce 打包命令...
         - 收集所有要包含在扩展中的文件
         - 生成扩展清单文件
         - 创建最终的 .vsix 包文件

> yuangs-vscode@1.0.5 package
> vsce package

 INFO  Files included in the VSIX:
...
 DONE  Packaged: /Users/ygs/yuangs-vscode/yuangs-vscode-1.0.5.vsix (23 files, 431.62 KB)

   ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
   ┃ 🎉 打包完成！VS Code 扩展包已成功创建                                  ┃
   ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

---

## 构建流程总结

### 完整流程图

```
1. 环境检查
   ├── 查找 Node.js 和 npm
   ├── 显示版本信息
   └── 检查 vsce 工具

2. 代码编译
   ├── 编译 AssemblyScript → WASM
   │   ├── debug 版本
   │   └── release 版本
   └── Webpack 打包
       ├── 编译 TypeScript
       ├── 模块捆绑
       ├── 代码优化
       └── 复制资源文件

3. 扩展打包
   ├── 验证 package.json
   ├── 执行 prepublish
   ├── 收集文件
   ├── 生成清单
   └── 创建 VSIX 包
```

### 构建产物对比

| 版本 | 文件大小 | 文件数量 | 备注 |
|------|---------|---------|------|
| 1.0.4 | 137,334 bytes (134 KB) | - | 旧版本 |
| 1.0.5 | 441,979 bytes (431 KB) | 23 files | 当前版本 |

**大小增加原因：**
- 包含了更多文档文件
- .ai/context.json (147.52 KB)
- 完整的编译产物
- Webview 资源

### 关键技术栈

1. **AssemblyScript (asc)**
   - 用于编译 WASM 模块
   - 实现高性能沙箱执行环境

2. **Webpack**
   - 模块打包工具
   - 代码压缩和优化
   - 依赖管理

3. **vsce (@vscode/vsce)**
   - VS Code 扩展打包工具
   - 生成 VSIX 安装包

### 使用建议

**开发阶段：**
```bash
./c --build-only
```
- 只编译代码，不打包
- 快速迭代，节省时间

**发布阶段：**
```bash
./c
```
- 完整编译和打包
- 生成可发布的 VSIX 包

**清理构建：**
```bash
./c --clean
```
- 删除旧的构建产物
- 确保干净构建环境

**测试扩展：**
```bash
code --install-extension yuangs-vscode-1.0.5.vsix
```

**调试扩展：**
- 在 VS Code 中按 F5
- 使用 Extension Development Host

---

## 构建时间分析

### 第一次执行（完整流程）
- AssemblyScript 编译: ~2秒
- Webpack 打包: ~29秒
- vsce 打包: ~13秒
- **总计: ~44秒**

### 第二次执行（prepublish 重复）
- AssemblyScript 编译: ~1.5秒
- Webpack 打包: ~10秒
- vsce 打包: ~11秒
- **总计: ~22.5秒**

**优化建议：**
- prepublish 重复编译可以优化
- 考虑使用增量构建
- 缓存依赖项减少重复编译

---

## 常见问题排查

### 1. vsce 未找到
```bash
npm install -g @vscode/vsce
```

### 2. Node.js 版本不兼容
确保使用 Node.js v22.x 或更高版本

### 3. AssemblyScript 编译失败
检查 `src/engine/agent/governance/sandbox/core.as.ts` 语法

### 4. Webpack 打包失败
检查依赖项是否正确安装：
```bash
npm install
```

### 5. vsce 打包失败
- 检查 package.json 必需字段
- 确保 README.md 和 LICENSE 存在
- 验证 publisher 字段已设置

---

## 结论

这个构建脚本实现了一个完整的自动化 CI/CD 流程，从环境检查到最终的 VSIX 包生成。通过将复杂的构建步骤封装在清晰的脚本中，大大简化了开发者的工作流程，确保了构建的一致性和可重复性。
