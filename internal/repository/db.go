package repository

import (
	"errors"
	"fmt"

	"github.com/zqr233qr/story-trim/internal/config"
	"github.com/zqr233qr/story-trim/internal/model"
	"github.com/zqr233qr/story-trim/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"
)

func NewDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Source), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.Book{},
		&model.Chapter{},
		&model.ChapterContent{},
		&model.Prompt{},
		&model.Task{},
		&model.TrimResult{},
		&model.UserProcessedChapter{},
		&model.ReadingHistory{},
		&model.User{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto migrate database: %w", err)
	}

	// 初始化数据库中的提示数据
	err = promptSeeder(db)
	if err != nil {
		return nil, fmt.Errorf("failed to seed prompts: %w", err)
	}

	logger.Info().Msg("Database connected and migrated successfully")
	return db, nil
}

// promptSeeder 初始化数据库中的提示数据
func promptSeeder(db *gorm.DB) error {
	prompts := []model.Prompt{
		{
			ID:             1,
			Name:           "轻量去水",
			Description:    "无损净化，去广告去乱码，保留原汁原味",
			IsSystem:       true,
			IsDefault:      true, // Default
			TargetRatioMin: 0.85, TargetRatioMax: 0.95,
			BoundaryRatioMin: 0.8, BoundaryRatioMax: 0.98,
			PromptContent: `### 执行细则 (Mode: Light)
1.  **对话红线 (绝对保护)**
    * **全量保留**：必须保留所有人物对话的字面完整性，包括方言、口癖和非标准语法，**严禁**修改对话内容。
    * **例外**：仅可剔除毫无意义的连续口吃（如“我...我...我...”改为“我...”）或单纯凑字数的语气词。

2.  **描写微调**
    * **修辞精简**：保留所有环境与动作描写，仅删除明显的同义重复（例如：“他心中充满了愤怒，怒火中烧，气得不行” → 保留“他心中怒火中烧”）。
    * **保留氛围**：对于渲染气氛的景物描写（如天气、光影），**一律保留**，不做删减。

3.  **连贯性要求**
    * 输出必须读起来像**精校版原文**，而不是缩写。任何删改都不应让读者察觉到文本经过了处理。
4.  **原词锁定**
	* 保留原文的用词习惯，哪怕是俚语、网络用语或语法不规范的句子，只要不影响核心理解，一律不予修正。`,
		},
		{
			ID:             2,
			Name:           "标准精简",
			Description:    "去脂保肉，删繁琐描写，提升阅读节奏",
			IsSystem:       true,
			IsDefault:      false,
			TargetRatioMin: 0.65, TargetRatioMax: 0.75,
			BoundaryRatioMin: 0.60, BoundaryRatioMax: 0.80,
			PromptContent: `### 执行细则 (Mode: Standard)
1.  **动作合并 (Action Merge)**
    * 将连续的琐碎分解动作合并为紧凑的单句。
    * *示例*：“他伸出手，握住了门把手，轻轻向下按压，推开了门。” → *处理为*：“他拧开门把手推门而入。
	* 拒绝成语化：合并动作时，使用动词直描。
		❌ 错误：他大快朵颐。
		✅ 正确：他大口吃肉。（如果原文是写吃肉）”

2.  **对话提炼**
    * **保留核心**：保留推动剧情、展现性格、揭示伏笔的对话。
    * **删减废话**：删除两人之间单纯的寒暄、无信息量的互相吹捧、重复的质问（如A问“真的吗？”，B答“真的”，A又问“确定？”）。

3.  **环境与心理降维**
    * **心理活动**：将大段的内心独白（超过3句的）压缩为1句精准的状态描述。
    * **环境描写**：除非该环境对战斗或剧情有物理影响（如地形障碍），否则删除单纯的风景描写。`,
		},
		{
			ID:             3,
			Name:           "极简速读",
			Description:    "只看干货，保留核心冲突，剧情极速推进",
			IsSystem:       true,
			IsDefault:      false,
			TargetRatioMin: 0.35, TargetRatioMax: 0.45,
			BoundaryRatioMin: 0.3, BoundaryRatioMax: 0.5,
			PromptContent: `### 执行细则 (Mode: Speed)
1.  **剧情骨架化 (Skeleton Only)**
    * **只写主干**：严格遵循“谁（Who）+ 在哪里（Where）+ 做了什么（Action）+ 结果（Result）”的结构。
    * **剔除装饰**：彻底删除所有修饰性形容词（如“凄美的”、“震耳欲聋的”）、环境渲染和次要人物的反应描写。
	* **说明文口吻**：使用冷淡、客观的说明文语气。不需要渲染战斗的紧张感，只需要交代谁赢了。不要使用感叹号。

2.  **对话转述策略**
    * **非必要不引用**：除了绝妙的装逼打脸台词或关键情报外，将大部分对话转化为叙述性文本。
    * *示例*：原书两人争吵了500字 → *处理为*：“两人因战利品分配不均爆发激烈争吵，最终决定平分。”

3.  **战斗与升级**
    * **战斗简化**：不描写一招一式的具体过程，直接描写关键技能的释放与战斗结果（谁赢了，谁受伤了，掉了什么装备）。
    * **数据简化**：不要列出详细的属性面板，直接说“全属性大幅提升”或“获得了神器XXX”。`,
		},
		{
			ID:             4,
			Name:           "总结摘要",
			Description:    "蒙太奇式快进，极速看完本章核心剧情",
			IsSystem:       true,
			IsDefault:      false,
			TargetRatioMin: 0.10, TargetRatioMax: 0.15,
			BoundaryRatioMin: 0.07, BoundaryRatioMax: 0.2,
			PromptContent: `### 执行细则 (Mode: Plot Synopsis)
**核心宗旨**：**彻底抛弃小说笔法**。像撰写“维基百科剧情简介”或“新闻通稿”一样，只陈述**事实（Fact）**，删除所有**情绪（Emotion）**和**感知（Sensation）**。

1.  **绝对禁令 (Zero Tolerance)**
    * **禁对话**：**严禁出现任何双引号（""）**。必须将所有对话转化为行动总结（例如：A嘲讽了B，C安慰了B）。
    * **禁描写**：**删除所有感官描写**。不要写动作细节（如“握紧拳头”）、不要写表情（如“面露苦涩”）、不要写环境（如“刺眼的光芒”）。
    * **禁抒情**：不要写心理活动（如“嘲讽像针扎一样”），只写心理状态（如“他感到痛苦”）。

2.  **情报提取逻辑 (Information Extraction)**
    * 请将本章压缩为 **3-5 个逻辑紧密的句子**，组成 1-2 个自然段。
    * 只保留：**谁（Who） + 做了什么/遭遇了什么（Event） + 结果如何（Result）**。

3.  **示例对照 (Few-Shot)**
    * ❌ 错误（小说写法）：萧炎看着石碑，痛苦地握紧了拳头，指甲刺破了手掌。众人大笑道：“废物！”
    * ✅ 正确（速报写法）：萧炎测验结果仅为三段，因实力大跌而遭到族人公开嘲讽。`,
		},
	}

	for _, p := range prompts {
		// Use ID as constraint for seeding
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&p).Error; err != nil {
			return err
		}
	}

	return nil
}

// FirstRecodeIgnoreError 获取第一条记录，忽略错误
func FirstRecodeIgnoreError(db *gorm.DB, dest interface{}) (bool, error) {
	if err := db.First(&dest).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistWithoutObject 检查表中是否存在指定的记录，不返回记录本身
func ExistWithoutObject(db *gorm.DB) (bool, error) {
	var count int64
	err := db.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
