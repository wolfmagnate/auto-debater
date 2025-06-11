package main

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/wolfmagnate/auto_debater/domain"
	"github.com/wolfmagnate/auto_debater/rebuttal_analyzer"
)

func TestMain(t *testing.T) {
	// カレントディレクトリの".env"を読み込んで環境変数に設定するコードを書いてください
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	main()
}

func TestRebuttal(t *testing.T) {
	// output.jsonから読み取りdomain.NewDebateGraphFromJSONを利用する
	if err := godotenv.Load(); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Read the debate graph from output.json
	jsonBytes, err := os.ReadFile("output.json")
	if err != nil {
		t.Fatalf("Failed to read output.json: %v", err)
	}

	debateGraph, err := domain.NewDebateGraphFromJSON(string(jsonBytes))
	if err != nil {
		t.Fatalf("Failed to create DebateGraph from JSON: %v", err)
	}

	rebuttalAnalyzer, err := rebuttal_analyzer.CreateRebuttalAnalyzer()
	if err != nil {
		t.Fatalf("Failed to create RebuttalAnalyzer: %v", err)
	}

	rebuttal := `
反対側の主張は、国政選挙におけるインターネット投票導入に反対し、現行の投票所方式を維持すべきだという立場を強固に論証するものです。しかし、その主張は、現行制度が抱える課題から目を背け、テクノロジーの可能性を過小評価していると言わざるを得ません。以下に、その論理構造に対する反論を述べます。

第一に、現行制度が「民主主義の信頼性と公正性を維持」してきたという主張は、深刻化する投票率の低下という現実を前に、その正当性が揺らいでいます。特に若年層や働き世代の投票率の低さは、もはや看過できないレベルに達しており、一部の世代や層の声だけが過剰に政治に反映される「サイレントマジョリティ」の増大を招いています。これは、民主主義の根幹である民意の正確な反映という観点から、信頼性・公正性が損なわれている状況と言えます。インターネット投票は、物理的・時間的な制約から投票を諦めていた人々の参加を促し、この歪みを是正する極めて有効な手段です。

第二に、「サイバー攻撃のリスク」や「個人情報漏洩のリスク」を絶対的な障壁と見なすのは、技術の進歩を無視した議論です。現代社会では、金融取引や重要な行政手続きの多くが、既に高度なセキュリティ対策のもとでオンライン化されています。選挙システムも同様に、ブロックチェーン技術やマイナンバーカードと連携した厳格な多要素認証などを活用することで、不正アクセスやなりすましを防ぎ、データの秘匿性と正確性を担保することは十分に可能です。エストニアをはじめとする電子投票先進国の事例は、適切な技術的・制度的設計により、リスクは管理可能であることを示唆しています。リスクをゼロにすることに固執し、変化を拒むことは、より多くの民意を汲み取るという民主主義の発展機会を逸することに繋がります。

第三に、「高齢者や情報弱者の投票機会が失われる」という懸念は重要ですが、それはインターネット投票を導入しない理由にはなりません。むしろ、社会全体のデジタル化を推進する中で、誰一人取り残さないためのサポート体制を構築する好機と捉えるべきです。導入に際しては、現行の投票所方式と併用し、有権者が選択できる環境を整えるべきです。公共施設での操作サポート窓口の設置や、移動式のデジタル投票支援など、多様な選択肢を用意することで、デジタル格差の問題は克服可能です。

最後に、「投票所へ足を運ぶ行為が選挙への意識を高める」という主張は、多分に情緒的であり、その効果は限定的です。選挙への意識は、日々の生活の中で政治の重要性を実感し、政策情報を得て自ら考えるプロセスを通じて醸成されるものです。インターネット投票の導入は、オンラインでの政策議論の活発化や、候補者情報へのアクセシビリティ向上を促し、より本質的な政治参加意識の涵養に貢献する可能性を秘めています。

結論として、現行制度の維持は、民主主義の停滞を容認することに他なりません。リスクを適切に管理し、テクノロジーの恩恵を最大限に活用することで、より多くの国民が参加し、多様な民意が反映される、より成熟した民主主義を実現するために、国政選挙へのインターネット投票導入に向けた前向きな議論を始めるべきです。
`

	err = rebuttalAnalyzer.AnalyzeRebuttal(context.Background(), debateGraph, rebuttal)
	if err != nil {
		t.Fatalf("反論の分析に失敗しました: %v", err)
	}

	// 分析後のグラフを
}
