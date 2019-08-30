// ==== システムの構成要素 ====
// System: 分散ストレージelton2のプログラムのこと。
// Subsystem: メタデータ管理、ストレージなど大きな機能を提供しているプログラムのこと。
//            1つのsubsystemは、複数のserviceを持つ。
// Service: subsystemの一部。1つのserviceは、単一の機能を提供するプログラムのこと。
// --------
// Process: OS用語でのプロセスとほぼ同義。1プロセスは、１つまたは複数のserverを実行することができる。
// Server: 1つのserviceを動かす。1つのlistenしているアドレスが必要。
