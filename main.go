package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/wolfmagnate/auto_debater/debate_graph_creator"
	"github.com/wolfmagnate/auto_debater/logic_graph_creator"
)

func main() {
	causeFinder, err := logic_graph_creator.CreateCauseFinder()
	if err != nil {
		log.Fatalf("ImpactAnalyzerの作成に失敗しました: %v", err)
	}

	newArgumentFinder, err := logic_graph_creator.CreateNewArgumentFinder()
	if err != nil {
		log.Fatalf("NewArgumentFinderの作成に失敗しました: %v", err)
	}

	logicGraphCompleter := &logic_graph_creator.LogicGraphCompleter{
		CauseFinder:       causeFinder,
		NewArgumentFinder: newArgumentFinder,
	}

	basicStructureAnalyzer, err := logic_graph_creator.CreateBasicStructureAnalyzer()
	if err != nil {
		log.Fatalf("BasicStructureAnalyzerの作成に失敗しました: %v", err)
	}

	impactAnalyzer, err := logic_graph_creator.CreateImpactAnalyzer()
	if err != nil {
		log.Fatalf("ImpactAnalyzerの作成に失敗しました: %v", err)
	}

	benefitHarmConverter, err := logic_graph_creator.CreateBenefitHarmConverter()
	if err != nil {
		log.Fatalf("BenefitHarmConverterの作成に失敗しました: %v", err)
	}

	logicGraphCreator := &logic_graph_creator.LogicGraphCreator{
		BasicStructureAnalyzer: basicStructureAnalyzer,
		ImpactAnalyzer:         impactAnalyzer,
		BenefitHarmConverter:   benefitHarmConverter,
		LogicGraphCompleter:    logicGraphCompleter,
	}

	document := `
日本における国政選挙へのインターネット投票導入論に対し、我々は断固として反対の立場を表明します。現行の投票所における記名投票・秘密投票の原則に基づいた制度は、我が国の民主主義を長年にわたり支えてきた基盤であり、その信頼性と公正性を軽々に損なうべきではありません。インターネット投票の導入は、一見利便性が向上するように見えますが、それと引き換えに失うものが大きすぎると考えます。

第一に、投票の公正性および秘密保持の困難性です。インターネットを介した投票では、技術的にどれほど対策を講じても、第三者による不正アクセスや投票内容の覗き見、あるいは有権者本人への圧力といったリスクを完全に排除することは極めて困難です。家庭内や特定のコミュニティ内で、他者の意思が投票行動に不当に介入する危険性も高まります。これに対し、現行の投票所方式は、厳格な本人確認と独立した記載場所の提供により、個々人の自由な意思に基づく秘密投票を最大限に保障しています。

第二に、サイバーセキュリティ上の深刻な脅威です。国家の根幹をなす選挙システムがサイバー攻撃の標的となれば、投票結果の改ざん、投票プロセスの妨害、大規模なシステムダウンなど、民主主義の根幹を揺るがす事態を招きかねません。また、投票者の個人情報や投票行動に関するデータが漏洩するリスクも無視できません。一度失われた選挙への信頼を取り戻すことは容易ではなく、その代償は計り知れません。

第三に、投票参加における機会の不均等と、投票行動の質の低下です。インターネット環境やデジタル機器の操作スキルには個人差があり、特に高齢者や情報弱者と呼ばれる層にとって、インターネット投票は新たな障壁となり得ます。これは投票率向上どころか、かえって特定層の投票機会を奪う結果になりかねません。また、投票所へ足を運ぶという行為が持つ、選挙への意識高揚や社会参加の実感を軽視すべきではありません。利便性のみを追求するあまり、一票の重みや熟慮のプロセスが希薄化することも懸念されます。

以上の理由から、我々は国政選挙へのインターネット投票導入に強く反対します。現行制度の信頼性を堅持しつつ、期日前投票の充実や移動投票所の拡充など、全ての人々が公平かつ安全に投票できる環境整備こそ、優先して取り組むべき課題であると考えます。安易なデジタル化が、民主主義の根幹を蝕むことのないよう、慎重な議論を求めます。
	`

	logicGraph, err := logicGraphCreator.CreateLogicGraph(context.Background(), document)
	if err != nil {
		log.Fatalf("ロジックグラフの作成に失敗しました: %v", err)
	}

	debateAnnotationCreator, err := debate_graph_creator.CreateDebateAnnotationCreator()
	if err != nil {
		log.Fatalf("DebateAnnotationCreatorの作成に失敗しました: %v", err)
	}

	documentSplitter, err := debate_graph_creator.CreateDocumentSplitter()
	if err != nil {
		log.Fatalf("DocumentSplitterの作成に失敗しました: %v", err)
	}

	debateGraphCreator := &debate_graph_creator.DebateGraphCreator{
		DebateAnnotationCreator: debateAnnotationCreator,
		DocumentSplitter:        documentSplitter,
	}

	debateGraph, err := debateGraphCreator.CreateDebateGraph(context.Background(), document, logicGraph)
	if err != nil {
		log.Fatalf("ディレクトリグラフの作成に失敗しました: %v", err)
	}

	jsonString, err := debateGraph.ToJSON()
	if err != nil {
		log.Fatalf("JSONに変換に失敗しました: %v", err)
	}

	file, err := os.Create("output.json")
	if err != nil {
		log.Fatalf("ファイルの作成に失敗しました: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(jsonString)
	if err != nil {
		log.Fatalf("ファイルへの書き込みに失敗しました: %v", err)
	}

	fmt.Println("Debate graph successfully written to output.json")

}
