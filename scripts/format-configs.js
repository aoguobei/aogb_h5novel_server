const fs = require("fs");
const path = require("path");

// 参考文件列表，这些文件不需要格式化
const REFERENCE_FILES = ["shiguang", "shuxiang", "yunchuan", "jiutian"];

class ConfigFormatter {
  constructor() {
    this.referenceFiles = REFERENCE_FILES;
  }

  // 格式化所有配置文件
  async formatConfigFiles(baseDir) {
    const configDirs = ["baseConfigs", "commonConfigs", "payConfigs", "uiConfigs"];
    
    for (const dir of configDirs) {
      const dirPath = path.join(baseDir, "src", "appConfig", dir);
      await this.formatDirectory(dirPath);
    }
  }

  // 格式化目录中的所有配置文件
  async formatDirectory(dirPath) {
    try {
      const entries = fs.readdirSync(dirPath);
      
      for (const entry of entries) {
        const fullPath = path.join(dirPath, entry);
        const stat = fs.statSync(fullPath);
        
        if (stat.isDirectory() || !entry.endsWith(".js")) {
          continue;
        }

        // 检查是否为参考文件
        const fileName = path.basename(entry, ".js");
        if (this.isReferenceFile(fileName)) {
          console.log(`⏭️  跳过参考文件: ${entry}`);
          continue;
        }

        await this.formatFile(fullPath);
      }
    } catch (error) {
      console.error(`❌ 格式化目录失败 ${dirPath}:`, error.message);
    }
  }

  // 检查是否为参考文件
  isReferenceFile(fileName) {
    return this.referenceFiles.includes(fileName);
  }

  // 格式化单个配置文件
  async formatFile(filePath) {
    try {
      console.log(`🔄 开始格式化文件: ${path.basename(filePath)}`);
      
      // 读取文件内容
      const content = fs.readFileSync(filePath, "utf8");
      
      // 解析配置
      const configData = this.parseConfigFile(content);
      
      // 格式化配置（按字典序排序）
      const formattedData = this.formatConfigData(configData);
      
      // 生成格式化后的内容
      const formattedContent = this.generateFormattedContent(formattedData);
      
      // 写入文件
      fs.writeFileSync(filePath, formattedContent, "utf8");
      
      console.log(`✅ 文件格式化完成: ${path.basename(filePath)}`);
    } catch (error) {
      console.error(`❌ 格式化文件失败 ${filePath}:`, error.message);
    }
  }

  // 解析配置文件
  parseConfigFile(content) {
    // 移除export default前缀
    if (content.startsWith("export default ")) {
      content = content.replace("export default ", "");
    }

    // 清理多余的逗号和格式问题
    content = this.cleanTrailingCommas(content);

    try {
      return JSON.parse(content);
    } catch (error) {
      console.log(`原始内容: ${content}`);
      throw new Error(`JSON解析失败: ${error.message}`);
    }
  }

  // 清理多余的逗号
  cleanTrailingCommas(content) {
    console.log(`开始清理内容，原始长度: ${content.length}`);
    
    // 将单引号替换为双引号，使其符合JSON格式
    content = content.replace(/'/g, '"');
    
    // 使用正则表达式移除对象末尾的多余逗号
    // 匹配模式：, 后面跟着任意空白字符，然后是 } 或 ]
    content = content.replace(/,(\s*[}\]])/g, "$1");
    
    // 移除行末的多余逗号
    const lines = content.split('\n');
    const finalLines = [];

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      const trimmedLine = line.trim();

      // 如果这行以逗号结尾，检查下一行
      if (trimmedLine.endsWith(',')) {
        if (i + 1 < lines.length) {
          const nextLine = lines[i + 1].trim();
          // 如果下一行是 } 或 ] 开头，移除逗号
          if (nextLine.startsWith('}') || nextLine.startsWith(']')) {
            finalLines.push(line.replace(/,\s*$/, ''));
            console.log(`移除行 ${i + 1} 的多余逗号`);
            continue;
          }
        }
      }

      finalLines.push(line);
    }

    // 重新组合内容
    content = finalLines.join('\n');
    
    // 最后检查：移除对象末尾的逗号
    content = content.replace(/,\s*}/g, '}');
    content = content.replace(/,\s*]/g, ']');
    
    console.log(`清理完成，处理后长度: ${content.length}`);
    return content;
  }

  // 格式化配置数据（按字典序排序）
  formatConfigData(configData) {
    const formattedData = {};

    // 对顶级键进行字典序排序
    const keys = Object.keys(configData).sort();

    // 格式化每个配置项
    for (const key of keys) {
      if (typeof configData[key] === "object" && configData[key] !== null) {
        formattedData[key] = this.formatConfigObject(configData[key]);
      } else {
        formattedData[key] = configData[key];
      }
    }

    return formattedData;
  }

  // 格式化配置对象（按字典序排序）
  formatConfigObject(config) {
    const formattedConfig = {};

    // 对对象的所有键进行字典序排序
    const keys = Object.keys(config).sort();
    
    for (const key of keys) {
      const value = config[key];
      // 递归处理嵌套对象
      if (typeof value === "object" && value !== null && !Array.isArray(value)) {
        formattedConfig[key] = this.formatConfigObject(value);
      } else {
        formattedConfig[key] = value;
      }
    }

    return formattedConfig;
  }

  // 生成格式化后的文件内容
  generateFormattedContent(configData) {
    const configJSON = JSON.stringify(configData, null, 2);
    return `export default ${configJSON}\n`;
  }
}

// 主函数
async function main() {
  const formatter = new ConfigFormatter();
  
  // 获取funNovel项目路径 - 修复路径问题
  const funNovelPath = path.join(__dirname, "..", "..", "funNovel");
  
  console.log("🚀 开始格式化配置文件...");
  console.log(`📁 项目路径: ${funNovelPath}`);
  
  try {
    await formatter.formatConfigFiles(funNovelPath);
    console.log("🎉 所有配置文件格式化完成！");
  } catch (error) {
    console.error("❌ 格式化过程中出现错误:", error.message);
    process.exit(1);
  }
}

// 如果直接运行此脚本
if (require.main === module) {
  main();
}

module.exports = ConfigFormatter;
