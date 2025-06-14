# タスク
与えられた論理構造グラフの中の特定の因果関係を最も効果的に強化するために人間がどのような調査を行うべきかを教えてください。

# 論理構造グラフの基本構造
## グラフの構造
与えられた論理構造グラフは、説得を行うための文章に対応します。説得では、現状維持の選択肢であるStatus Quoと積極的な改善策を行うAffirmative Planを比較する。比較では、Status Quoを前提とした世界とAffirmative Planを前提とした世界で発生する因果関係を分析し、最終的なメリット・デメリットを主張する。
例えば、原発再稼働というAffirmative Planに対して賛成の立場を取る場合、「原発再稼働」が「エネルギーの安定供給」を引き起こし、「エネルギーの安定供給」が「経済発展」を引き起こすという、「原発再稼働」を原因として引き起こされるメリットを主張し、「原発停止」が「火力発電依存」を引き起こし、「火力発電依存」が「CO2の排出増加による地球温暖化」を引き起こすという、「原発停止」を原因として引き起こされるデメリットを主張する。

論理構造グラフでは、ノードはこのような何らかの主張に対応し、エッジは因果関係に対応します。エッジは有向辺で、原因から結果に対して辺が引かれます。

## 因果関係
因果関係とは、原因が結果を引き起こす関係です。これは目的と手段の関係と勘違いしやすいため注意が必要です。
例えば「エネルギーの安定供給ができるから、原発再稼働をする」という文章では、「AだからB」というAが原因のように見える文章です。
しかし、実際はエネルギーの安定供給を達成すればそれによって原発再稼働が引き起こされるわけではありません。「エネルギーの安定供給という目標があるから、その達成のために原発再稼働という手段を取る」という意図の文章です。かならず、「AがBを引き起こす」という文章が論理的に自然な場合のみ因果関係であると判断してください。

与えられた論理構造グラフではStatus QuoとAffirmative Planが議論の究極的な前提（原因）となっており、最終的にメリット・デメリットが引き起こされる結果になっています。

## 重要性
説得のためには、メリット・デメリットが重要なものであることが必要です。そのため、論点の重要さ、深刻さを強調したり、より多くの人間が影響を受けると主張したりすることで、メリットが達成されることの大切さを強調し、デメリットを受けることの深刻さを強調します。
このために行う主張をノードの重要性(importance)と呼びます。

重要性の例：「地球温暖化は、異常気象による自然災害の激甚化、海面上昇による生活圏の喪失、食糧生産への打撃など、私たちの生存基盤そのものを脅かす地球規模の喫緊の課題です」（「地球温暖化の解決」の重要性）

## 独自性
説得力はStatus QuoとAffirmative Planの世界の差分から生まれます。したがって、特定の因果関係や主張が片方の世界でのみ発生することが重要です。
この差分を示すための主張をエッジの独自性(uniqueness)、ノードの独自性(uniqueness)と呼びます。

エッジの独自性の例：「原発は自然条件に依存する再生可能エネルギーや中東に集中して分布する火力発電と違って世界中に広く分布するウランから発電でき、地政学リスクを避けられます」（「原発再稼働」が「電力の安定供給」を引き起こす独自性）
ノードの独自性の例：原発再稼働反対議論において、特定の地域へのリスク集中というデメリットノードに対して「物理学的に原発事故時の放射線は近い場所から強くなります」

## 確実性
論理構造グラフではStatus QuoとAffirmative Planを原因、メリット・デメリットを結果とする因果関係が述べられます。
このグラフでは、因果関係が確実に成立することが重要です。原因が結果を引き起こす可能性が低い場合には、議論全体の信頼性が下がります。
そこで、具体例を出したり、なぜ原因が結果を引き起こしやすいかを説明します。このような主張をエッジの確実性(certainty)と呼びます。

確実性の例：「電力は全ての産業活動の基盤です。安価で安定した電力が供給されることで、企業は生産コストを予測しやすくなり、設備投資や研究開発といった将来への投資判断がしやすくなります」（「安定した電力供給」が「経済発展」を引き起こすことの確実性）

# 論理構造グラフ
この情報はあくまで参考となる文章です。

{{.DebateGraphJSON}}

# 強化する因果関係
このサブ論理構造グラフが本質的に強化したいものです。このグラフの最も根となるノードが強化対象の究極の原因であり、最も葉となるノードが究極の結果です。原因が結果を引き起こす可能性が高いことを説明するためのロジックを作成してください。

{{.TargetDebateGraphJSON}}

# 強化の方法
論理構造グラフを強化するための方法は以下のいずれかです。

## 中間ノードの追加
既存のグラフの原因と結果までの間に新しいノードを追加します。
例えば「円安」→「国民生活の悪化」という議論に対して中間ノード 「輸入製品価格の上昇」 を追加します。「円安によって輸入品が高くなり、それが生活を圧迫する」という、よりスムーズで理解しやすい論理の流れが完成します。
人間は調査によって中間的な原因があるのではないかという考察を行うことができます。あなたは「既存の因果関係に別の要因があるはずです、こういう調査をしてみませんか？実際の中間的な原因にはこのようなものがあるかもしれません、確めてみませんか」などと提案してください。

## エッジの強化
既存のエッジのcertaintyとuniquenessを強化します。

対象エッジ: 「エネルギーの安定供給」 → 「経済発展」
強化の種類 (Certainty): この因果関係の確実性を高める。
内容の例: 「安価で安定した電力は、工場の24時間稼動や大規模なデータセンターの運用を可能にし、企業の国際競争力を直接的に高めます。さらに、電力コストの低下は物価の安定に繋がり、個人消費を刺激するため、マクロ経済全体に好影響を与えます。」
人間は調査やインタビューによって本当に因果関係が成立するかを確めることができます。あなたは「このエッジの因果関係が不明です。具体的にはこういう手段を通じて確めてみませんか」などと提案してください。

強化の種類 (Uniqueness): この因果関係の独自性（Planの世界でのみ起こること）を強調する。AがBを引き起こすのはこの世界の場合だけだ。
内容の例: 「特に大規模な企業ほど長期的に大規模に投資をします。その中では安定性という要素が絶対必要不可欠です」
人間は自分の提案した事業計画だけが問題を解決し、現状とは違う大きな差があることを確めるための調査を行うことができます。同様に差分を評価するための具体的な手法を提案してください。

## ノードの強化
この強化は、ノードがユーザーにとってのメリット・デメリットである場合やノードの内容がユーザーが価値の対価として一定の金額を支払うという内容である場合にのみ有効です。次のような内容を確めるための人間が行える調査がある場合、そのやり方をノードでの強化提案を行ってください。

## ターゲット顧客は誰か？
あなたが「この人たちのために事業をやる」と断言できる顧客像は、具体的にどのような人々か？（年齢、性別、職業、価値観、ライフスタイルなど）論理構造グラフのStatus Quoで課題を抱えている人のうち実際に顧客になる人が不明確な場合は指摘してください。
## 市場規模は十分か？
課題を抱えている顧客セグメントは、あなたの事業が持続的に成長できるだけの十分な大きさを持っているか？（TAM/SAM/SOM）
## その価値は、顧客にとって本当に重要か？
あなたが「強み」だと思っていることは、顧客がお金を払ってでも手に入れたいものと一致しているか？Affirmative Planで分析されているメリットが最終的に顧客の「お金を支払う」という行動につながらないと感じた場合指摘してください。
## 誰が、何に対して、お金を払うのか？
ユーザー（利用者）とバイヤー（支払者）は同じか？価値を感じるポイントと、課金ポイントは一致しているか？論理構造グラフのAffirmative Planでお金を払う人間が不明確な場合指摘してください。
## 価格設定は妥当か？
提供価値と価格のバランスは取れているか？安すぎたり、高すぎたりしないか？与える価値について
## ユニットエコノミクスは成立しているか？
顧客一人あたりの生涯価値（LTV）は、顧客一人あたりの獲得コスト（CAC）を十分に上回っているか？

# 人間の特性
このツールによって提案した行動にしたがうのは人間です。したがって、論理的な内容（説得力があり筋が通ること）以上に、実際はどうなのか、予想もしていない論理構造はないか、一般的なAIのDeep Researchで調べられるような内容以上の特定顧客、市場、ユーザーに特化した内容を調べることができます。あなたの提案は、人間ならではの調査方法になるように提案してください。

# 分析の注意点
新しい内容を追加してください。例えば[再生可能エネルギーの導入が増加する] -> [CO2排出量が削減される]という論理構造について、既に中間ノードとして「化石燃料の割合が減少する」という内容を追加していた場合、類似した中間ノードを使っても説得力は増えません。ノードの強化、エッジの強化、新しい内容の追加をバランスよく組み合わせてください。すでにノードを追加している場合（強化する因果関係において究極の原因と結果以外の中間ノードがある場合）その他のやり方を模索してください。また、既にエッジの確実性を主張しているなら独自性を探すなどしてください。

論理的に当然の内容よりも主張を補足する知識や性質について、深い洞察や専門知識に基づく踏み込んだ理由付けをしてください。

もっとも有効な調査を2つ考えてください。

# 出力形式
次のGoの構造体に合わせてください。

```go
type TODOSuggestions struct {
    TODOs []EnhancementTODO `json:"todo"`
}

type EnhancementTODO struct {
    // 簡潔にこのTODOがどのようなものかを数十文字以内でまとめてタイトルを作ってください
    Title string `json:"title"`

	StrengthenEdge *StrengthenEdgePayload `json:"strengthen_edge,omitempty"`
    StrengthenNode *StrengthenNodePayload `json:"strengthen_node,omitempty"`
	InsertNode     *InsertNodePayload     `json:"insert_node,omitempty"`
}

// エッジの強化
type StrengthenEdgePayload struct {
	CauseArgument  string `json:"cause_argument"`
	EffectArgument string `json:"effect_argument"`
	// EnhancementType は "uniqueness" または "certainty" のいずれかです。
	EnhancementType string `json:"enhancement_type"`
	Content         string `json:"content"`
}

// ノードの強化
type StrengthenNodePayload struct {
    TargetArgument string `json:"target_argument"`
    // 具体的にPMFを達成するに値する課題とメリットなのか
    Content string `json:"content"`
}

// 特定のノードの間に中間的な原因があるのではないか、調べると有益なのではないかと考えられる場合にはこの提案をしてください
type InsertNodePayload struct {
	CauseArgument        string `json:"cause_argument"`
	EffectArgument       string `json:"effect_argument"`
	IntermediateArgument string `json:"intermediate_argument"`
}
```