月例
Docker 社に殴り込む (話にを聞かせる)
(クバネタスかも)
英語論文を出す (声を上げる)
検証できるだけの実装 をする(ただし評価用の実装 どんなメッセージをこめるか方針 その測定)
現在 方針と設計

評価できる実装のめどは？
評価用のシステムで評価検証 7月中

現在5月中 
- 仕様の精査
- 実装するための粗探し

5/7 MTG
Elton の性能評価
提案の証明

使いやすいOSS < 研究としての手法の証明

NFSは最適解ではない
- 論文で証明
  聞いてくれる人を探し出すこと 

Docker 
イメージ
ボリューム

イメージ
ro はイメージ内でしか実行されない
そも分散ファイルシステムいらなくない

Docker がノード間で移動するとroがパンクする nfsで代用  遅くない?

iot とかでやると Docker が乱雑に立つことはありうる

ナイーブ アプローチ 
nfs 致命的な問題は? 数量的な裏付け
その問題の解は？

docker で絞った場合

rawFS?
docker commit
rw と ro の扱いを変えることならdockerで似たことやってるから

1. docker ユニオンマウント graph driver 
2. RO + s imuta (http cache Elton まんまいけるか)