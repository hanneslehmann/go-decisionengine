/*
Rule	Client Type	On Deposit	Estimated Net Worth	Client Category
Type	String List	Number	String List
1	"Business"	< 100000	"High"	"High Value Business"
2	"Business"	>= 100000	Not ("High")	"High Value Business"
3	"Business"	< 100000	Not ("High")	"Business Standard"
4	"Private"	>= 20000	"High"	"Personal Wealth Management"
5	"Private"	>= 20000	Not ("High")	"Personal Wealth Management"
6	"Private"	< 2000	-	"Personal Standard"

Test: curl -XPOST  'localhost:8888/api/v1/decision' -d '{"clientType":"Business","deposit":"10001","estNetWorth":"High"}'

*/
package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "net/http"
    "log"
    "encoding/json"
    "io"
    "io/ioutil"
    "reflect"
//    "strconv"
)

type DecisionDataType struct {
  Input InputRulesType `json:"inputs"`
  Output OutputType `json:"outputs"`
}

type DecisionRowType struct {
   ID int `json:"id"`
   Data DecisionDataType `json:"data"`
}

type DecisionTableType struct {
  Rows []DecisionRowType `json:"rows"`
}

type InputValueType struct {
  Value string `json:"value"`
  Rule string `json:"rule"`
}

type InputRulesType struct {
  I1 InputValueType `json:"clientType"`
  I2 InputValueType `json:"deposit"`
  I3 InputValueType `json:"estNetWorth"`
}

type InputType struct {
  I1 string `json:"clientType"`
  I2 string `json:"deposit"`
  I3 string `json:"estNetWorth"`
}

type OutputType struct {
  O1 string `json:"category"`
}

type ResponseType struct {
  Status string `json:"status"`
  Response OutputType `json:"response"`
}

var DecisionTable DecisionTableType

func init() {
  valBusiness:=InputValueType{"Business","equals"}
  valPrivate:=InputValueType{"Private","equals"}
  valLt1:=InputValueType{"10000","<"}
  valgteq1:=InputValueType{"10000",">="}
  valgteq2:=InputValueType{"10000",">="}
  valhigh:=InputValueType{"High","equals"}
  valnothigh:=InputValueType{"High","not equals"}
  valEmpty:=InputValueType{"",""}
  DecisionTable = DecisionTableType{
    []DecisionRowType{
      {1, DecisionDataType{InputRulesType{valBusiness, valLt1,valhigh},OutputType{"High Value Business"}}},
      {2, DecisionDataType{InputRulesType{valBusiness, valgteq1,valnothigh},OutputType{"High Value Business"}}},
      {3, DecisionDataType{InputRulesType{valBusiness, valLt1,valnothigh},OutputType{"Business Standard"}}},
      {4, DecisionDataType{InputRulesType{valPrivate, valgteq2,valhigh},OutputType{"Personal Wealth Management"}}},
      {5, DecisionDataType{InputRulesType{valPrivate, valgteq2,valnothigh},OutputType{"Personal Wealth Management"}}},
      {6, DecisionDataType{InputRulesType{valPrivate, valgteq2,valEmpty},OutputType{"Personal Standard"}}},
  },
}
}

func hitValue(input interface{}, rule interface{}) bool {
  ret:=false
  ivt:=rule.(InputValueType)
  fmt.Println("Checking ", input," against ", ivt.Value, "with rule", ivt.Rule)
  return ret
}

func (i InputRulesType) hitRule(input InputType) bool {
  ret:=false
  // get the fields of struct
  r := reflect.ValueOf(i)
  rules := make([]interface{}, r.NumField())
  v := reflect.ValueOf(input)
  values := make([]interface{}, v.NumField())
  for i := 0; i < v.NumField(); i++ {
        values[i] = v.Field(i).Interface()
  }
  for i := 0; i < r.NumField(); i++ {
        rules[i] = r.Field(i).Interface()
        ret=hitValue(v.Field(i).Interface(), rules[i])
  }
  return ret
}

func (d DecisionTableType) makeDecision(i InputType) OutputType {
  ret:=OutputType{"Undefined"}
  for _,y := range d.Rows {
     fmt.Println(y.Data.Input, "  ->", y.Data.Input.hitRule(i))
  }

  return ret
}

func handlePost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
  var data InputType
  body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
  if err != nil {
      panic(err)
  }
  if err := r.Body.Close(); err != nil {
      panic(err)
  }
  if err := json.Unmarshal(body, &data); err != nil {
      w.Header().Set("Content-Type", "application/json; charset=UTF-8")
      w.WriteHeader(422) // unprocessable entity
      if err := json.NewEncoder(w).Encode(err); err != nil {
          panic(err)
      }
  }
  w.Header().Set("Content-Type", "application/json; charset=UTF-8")
  w.WriteHeader(http.StatusCreated)
  decision:=DecisionTable.makeDecision(data)
  resp, _ := json.Marshal(ResponseType{"OK",decision})
  w.Write(resp)
  b, _ := json.Marshal(data)
  fmt.Println(string(b))
}

func main() {
  router := httprouter.New()
  b, _ := json.Marshal(DecisionTable)
  fmt.Println(string(b))
  router.POST("/api/v1/decision", handlePost)
  log.Fatal(http.ListenAndServe(":8888", router))
}
