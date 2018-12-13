tStart:=time.Now()
proverCommitMap = make(map[string][2]*big.Int)
proverSignCommitMap = make(map[string][2]*big.Int)
// step1:本人计算证明人身份验证的commit
cp := cryptocenter.CalcCommitmentParamsOfLockout()
// step2:本人计算证明人身份验证的zkp
zkpi1 := cryptocenter.CalcZeroKnowledgeProofI1ParamsOfLockout(cp)
// step3:广播身份验证的数据让其他公证人验证
var passedProvers= 0
for _, proverServe := range config.ProverServes {
   //urlPath := "http://" + proverServe + *config.HttpBindAddr + "/cct/lockout_req_check"
   urlPath := "http://127.0.0.1:" + proverServe + "/cct/lockout_req_check"
   checkResult := false
   cz := &broadcastDataOfLockout{cp, zkpi1}
   _, err := common.MakeRequest("POST", urlPath, &cz, &checkResult)
   if err != nil {
      logrus.Error(err)
   }
   if checkResult == true {
      passedProvers++
   }
}
if passedProvers != len(config.ProverServes) {
   proverCommitMap = nil
   return util.JSONResponse{
      Code: http.StatusNotAcceptable,
      JSON: "[LOCK-OUT]Sorry,there is a spy in the provers,and we will stop work for you,bye-bye!",
   }
}
// step4:收集U,V(广播通知其他证明人计算CommitParamsOfLockout)
for _, proverServe := range config.ProverServes {
   //urlPath := "http://" + proverServe + *config.HttpBindAddr + "/cct/lockout_notify_calc"
   urlPath := "http://127.0.0.1:" + proverServe + "/cct/lockout_notify_calc"
   //notifyResult := &cryptocenter.CommitParamsOfLockout{}
   notifyResult := new([2]*big.Int)
   res, err := common.MakeRequest("POST", urlPath, nil, &notifyResult)
   if err != nil {
      logrus.Error(err)
   }
   if res == nil {
      proverCommitMap = nil
      return util.JSONResponse{
         Code: http.StatusNotAcceptable,
         JSON: "[LOCK-OUT]Sorry,the prover[" + proverServe + "] missing,bye-bye!",
      }
   }
   proverCommitMap[proverServe] = *notifyResult
}
// step5:计算证明人阈值数量
//var allPartnerProve []*cryptocenter.CommitParamsOfLockout
var allPartnerProve [][2]*big.Int
for proverIp, data := range proverCommitMap {
   for _, ipx := range config.ProverServes {
      if proverIp == ipx {
         allPartnerProve = append(allPartnerProve, data)
      }
   }
}
// step6:本证明人合成u、v
u := cp.CalcSyntheticU(allPartnerProve)
v := cp.CalcSyntheticV(allPartnerProve)
// step7:本证明人计算我的sign-commit
csp := cryptocenter.CalcCommitmentSignParamsOfLockout(u, v)
// step8:本证明人计算我sign-zkpi2
zkpi2 := cryptocenter.CalcZeroKnowledgeProofI2SignParamsOfLockout(u, csp)
// step9:广播上述签名的数据让其他公证人验证
var passedProversSign= 0
for _, proverServe := range config.ProverServes {
   //urlPath := "http://" + proverServe + *config.HttpBindAddr + "/cct/lockout_req_sign_check"
   urlPath := "http://127.0.0.1:" + proverServe + "/cct/lockout_req_sign_check"
   checkResult := false
   scz := &broadcastSingDataOfLockout{csp, zkpi2, u}
   _, err := common.MakeRequest("POST", urlPath, &scz, &checkResult)
   if err != nil {
      logrus.Error(err)
   }
   if checkResult == true {
      passedProversSign++
   }
}
if passedProversSign != len(config.ProverServes) {
   proverCommitMap = nil
   return util.JSONResponse{
      Code: http.StatusNotAcceptable,
      JSON: "[LOCK-OUT]Sorry,there is a spy in the provers,and we will stop work for you,bye-bye!",
   }
}
// step10:收集w,r(广播通知其他证明人计算CommitSignParamsOfLockout)
for _, proverServe := range config.ProverServes {
   //urlPath := "http://" + proverServe + *config.HttpBindAddr + "/cct/lockout_notify_sign_calc"
   urlPath := "http://127.0.0.1:" + proverServe + "/cct/lockout_notify_sign_calc"
   //notifyResult := &cryptocenter.CommitSignParamsOfLockout{}
   notifyResult := new([2]*big.Int)
   uvx := uv{u, v}
   res, err := common.MakeRequest("POST", urlPath, &uvx, &notifyResult)
   if err != nil {
      logrus.Error(err)
   }
   if res == nil {
      proverSignCommitMap = nil
      return util.JSONResponse{
         Code: http.StatusNotAcceptable,
         JSON: "[LOCK-OUT]Sorry,the prover[" + proverServe + "] missing,bye-bye!",
      }
   }
   fmt.Println("收到来自[", proverServe, "]秘密0:", notifyResult[0])
   fmt.Println("收到来自[", proverServe, "]秘密1:", notifyResult[1])
   proverSignCommitMap[proverServe] = *notifyResult
}
// step11:计算证明人阈值数量
//var allPartnerSignProve []*cryptocenter.CommitSignParamsOfLockout
var allPartnerSignProve [][2]*big.Int
for proverIp, data := range proverSignCommitMap {
   for _, ipx := range config.ProverServes {
      if proverIp == ipx {
         allPartnerSignProve = append(allPartnerSignProve, data)
      }
   }
}
// step12:本证明人计算w、r
w := csp.CalcSyntheticW(allPartnerSignProve)
rx, ry := csp.CalcSyntheticR(allPartnerSignProve)
//计算和校验签名
message := "88888"
signature, pkx, pky := cryptocenter.CalcSignature(w, rx, ry, u, v, message)
fmt.Println("w:", w.BitLen(), ",", w)
fmt.Println("rx:", w.BitLen(), ",", rx)
fmt.Println("ry:", w.BitLen(), ",", ry)
fmt.Println("u:", w.BitLen(), ",", u)
fmt.Println("v:", w.BitLen(), ",", v)

logrus.Info("R:", signature.GetR())
logrus.Info("S:", signature.GetS())
logrus.Info("V:", signature.GetRecoveryParam())
logrus.Info("检查公钥x(32):", pkx.Bytes())
logrus.Info("检查公钥x(32):", pky.Bytes())
logrus.Info("时间差",time.Now().Sub(tStart).Seconds())
if signature != nil {
   if signature.Verify(message, pkx, pky) {
      return util.JSONResponse{
         Code: http.StatusOK,
         JSON: "[LOCK-OUT]Signature verified PASSED",
      }
   } else {
      return util.JSONResponse{
         Code: http.StatusNotFound,
         JSON: "[LOCK-OUT]Signature verified NOT-PASSED",
      }
   }
}
return util.JSONResponse{
   Code: http.StatusOK,
   JSON: "[LOCK-OUT]Signature verified NOT-PASSED,signature is null",
}



// step1:本人计算证明人身份验证的commit和zkp，
 // step1:本人计算证明人身份验证的commit和zkp，
 // step2:广播身份验证的数据让其他公证人验证
 // step4:收集U,V(其他人commit的结果)
 // step5:校验计算证明人阈值数量
 // step6:本证明人合成u、v
 // step7:本证明人计算我的sign-commit，同时广播通知其他证明人计算CommitSignParamsOfLockout
 // step8:本证明人计算我sign-zkpi2
 // step9:广播上述签名的数据让其他公证人验证
 // step10:收集w,r()
 // step11:计算证明人阈值数量
 // step12:本证明人计算w、r，
 // step13:计算和校验签名，同时通知其他人计算签名和签名校验，广播比对