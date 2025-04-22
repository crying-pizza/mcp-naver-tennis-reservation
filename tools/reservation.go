package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func GetAvailableTimeSlotHandler() {

	ctx := context.Background()
	request := mcp.CallToolRequest{
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Arguments: map[string]interface{}{
				"courtName": "양재",
				"year": 2025,
				"month": 4,
			},
		},
	}
	getAvailableTimeSlotHandler(ctx, request)
}

func GetAvailableTimeSlot() (mcp.Tool, server.ToolHandlerFunc) {
	return getAvailableTimeSlotTool(), getAvailableTimeSlotHandler
}

func getAvailableTimeSlotTool() mcp.Tool {
	return mcp.NewTool(
		"getAvailableTimeSlot",
		mcp.WithString("courtName",
			mcp.Required(),
			mcp.Description("Court name. Yangjae/Maeheon/Naegok"),
		),
		mcp.WithString("year",
			mcp.Required(),
			mcp.Description("year of the date"),
		),
		mcp.WithString("month",
			mcp.Required(),
			mcp.Description("month of the date"),
		),
	)
}

// func getAvailableTimeSlotTool() mcp.Tool {
// 	return mcp.NewTool(
// 		"getavailable",
// 		mcp.WithString("courtName",
// 			mcp.Required(),
// 			mcp.Description("Court name"),
// 		),
// 		mcp.WithString("year",
// 			mcp.Required(),
// 			mcp.Description("year of the date"),
// 		),
// 		mcp.WithString("month",
// 			mcp.Required(),
// 			mcp.Description("month of the date"),
// 		),
// 	)
// }

func createGraphQLUrl(opName string) string {
    return "https://booking.naver.com/graphql?opName=$opName"
}

func getBizItemIdMap(businessId, year, month int) (map[int]string, error) {
	payloadBizItems := map[string]interface{}{
		"operationName": "bizItems",
		"variables": map[string]interface{}{
			"withTypeValues":      false,
			"withReviewStat":      false,
			"withBookedCount":     false,
			"withClosedBizItem":   false,
			"input": map[string]interface{}{
				"businessId": strconv.Itoa(businessId),
				"lang":       "ko",
				"projections": "RESOURCE",
			},
		},
		"query": "query bizItems($input: BizItemsParams, $withTypeValues: Boolean = false, $withReviewStat: Boolean = false, $withBookedCount: Boolean = false) {\n  bizItems(input: $input) {\n    ...BizItemFragment\n    __typename\n  }\n}\n\nfragment BizItemFragment on BizItem {\n  id\n  agencyKey\n  businessId\n  bizItemId\n  bizItemType\n  name\n  desc\n  phone\n  stock\n  price\n  addressJson\n  startDate\n  endDate\n  refundDate\n  availableStartDate\n  bookingConfirmCode\n  bookingTimeUnitCode\n  isPeriodFixed\n  isOnsitePayment\n  isClosedBooking\n  isClosedBookingUser\n  isImp\n  minBookingCount\n  maxBookingCount\n  minBookingTime\n  maxBookingTime\n  extraFeeSettingJson\n  bookableSettingJson\n  bookingCountSettingJson\n  paymentSettingJson\n  bizItemSubType\n  priceByDates\n  websiteUrl\n  discountCardCode\n  customFormJson\n  optionCategoryMappings\n  bizItemCategoryId\n  additionalPropertyJson {\n    ageRatingSetting\n    openingHoursSetting\n    runningTime\n    parkingInfoSetting\n    ticketingTypeSetting\n    accommodationAdditionalProperty\n    arrangementCountSetting {\n      isUsingHeadCount\n      minHeadCount\n      maxHeadCount\n      __typename\n    }\n    bizItemCategorySpecificSetting {\n      instructorName\n      gatheringPlaceAddress\n      bizItemCategoryInfoMapping\n      bizItemCategoryInfo {\n        type\n        option\n        categoryInfoId\n        __typename\n      }\n      __typename\n    }\n    __typename\n  }\n  bookingCountType\n  isRequiringBookingOption\n  bookingUseGuideJson {\n    type\n    content\n    __typename\n  }\n  todayDealRate\n  extraDescJson\n  bookingPrecautionJson\n  isSeatUsed\n  isNPayUsed\n  isDeducted\n  isImpStock\n  isGreenTicket\n  orderSettingJson\n  isRobotDeliveryAvailable\n  bizItemAmenityJson {\n    amenityCode\n    amenityCategory\n    __typename\n  }\n  resources {\n    resourceUrl\n    __typename\n  }\n  bizItemResources {\n    resourceUrl\n    bizItemResourceSeq\n    bizItemId\n    order\n    resourceTypeCode\n    regDateTime\n    __typename\n  }\n  totalBookedCount @include(if: $withBookedCount)\n  currentDateTime @include(if: $withBookedCount)\n  reviewStatDetails @include(if: $withReviewStat) {\n    totalCount\n    avgRating\n    __typename\n  }\n  ...BizItemTypeValues @include(if: $withTypeValues)\n  ...MinMaxPrice\n  __typename\n}\n\nfragment BizItemTypeValues on BizItem {\n  typeValues {\n    bizItemId\n    code\n    codeValue\n    __typename\n  }\n  __typename\n}\n\nfragment MinMaxPrice on BizItem {\n  minMaxPrice {\n    minPrice\n    minNormalPrice\n    maxPrice\n    maxNormalPrice\n    isSinglePrice\n    discountRate {\n      min\n      max\n      __typename\n    }\n    __typename\n  }\n  __typename\n}",
	}

	// fmt.Println(payloadBizItems)

	payloadBytes, err := json.Marshal(payloadBizItems)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", createGraphQLUrl("bizItems"), bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, err
	}

	bizItemIdMap := make(map[int]string)

	monthStr := strconv.Itoa(month) + "월"
	data := result["data"].(map[string]interface{})
	bizItems := data["bizItems"].([]interface{})
	for _, item := range bizItems {
		it := item.(map[string]interface{})
		name := it["name"].(string)
		if len(name) >= len(monthStr) && name[:len(monthStr)] == monthStr {
			bizItemIdStr, ok := it["bizItemId"].(string)
			if !ok {
				return nil, fmt.Errorf("bizItemId가 string 타입이 아닙니다: %v", it["bizItemId"])
			}

			bizItemId, err := strconv.Atoi(bizItemIdStr)
			if err != nil {
				return nil, fmt.Errorf("bizItemId 문자열을 int로 변환 실패: %v", err)
			}

			bizItemIdMap[bizItemId] = name
		}
	}

	return bizItemIdMap, nil
}

func getSchedule(businessTypeId, businessId int, firstDay, lastDay string, bizItemId int) ([]interface{}, error) {
	payloadMap := map[string]interface{}{
		"operationName": "schedule",
		"variables": map[string]interface{}{
			"scheduleParams": map[string]interface{}{
				"businessTypeId":          businessTypeId,
				"businessId":              fmt.Sprintf("%d", businessId),
				"bizItemId":               fmt.Sprintf("%d", bizItemId),
				"startDateTime":           firstDay + "T00:00:00",
				"endDateTime":             lastDay + "T23:59:59",
				"fixedTime":               true,
				"includesHolidaySchedules": true,
			},
		},
		"query": `query schedule($scheduleParams: ScheduleParams) {
  schedule(input: $scheduleParams) {
    bizItemSchedule {
      hourly {
        isBusinessDay
        isSaleDay
        isUnitBusinessDay
        isUnitSaleDay
        isHoliday
        unitStock
        unitBookingCount
        bookingCount
        stock
        duration
        minBookingCount
        maxBookingCount
        unitStartTime
        unitStartDateTime
        slotId
        prices {
          isDefault
          priceId
          agencyKey
          slotId
          scheduleId
          name
          priceTypeCode
          price
          normalPrice
          desc
          order
          groupName
          groupOrder
          bookingCount
          isImp
          saleStartDateTime
          saleEndDateTime
          __typename
        }
        __typename
      }
      __typename
    }
    __typename
  }
}`,
	}

	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", createGraphQLUrl("schedule"), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		log.Printf("에러 발생: %d\n에러 내용: %s\n", res.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("API 요청 실패")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, err
	}

	// 결과 파싱
	hourly := []interface{}{}
	data := parsed["data"].(map[string]interface{})
	schedule := data["schedule"].(map[string]interface{})
	bizItemSchedule := schedule["bizItemSchedule"].(map[string]interface{})
	if h, ok := bizItemSchedule["hourly"].([]interface{}); ok {
		hourly = h
	}

	return hourly, nil
	// result := map[int][]interface{}{
	// 	bizItemId: hourly,
	// }
	// return result, nil
}

// type result struct {
// 	id   int
// 	data []interface{}
// }

// func getSchedulesAsync(businessTypeId, businessId int, firstDay, lastDay string, bizItemIdList []int) map[int][]interface{} {
// 	var wg sync.WaitGroup
// 	resultCh := make(chan result, len(bizItemIdList))

// 	for _, id := range bizItemIdList {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			res, err := getSchedule(businessTypeId, businessId, firstDay, lastDay, id)
// 			if err != nil {
// 				log.Printf("error for bizItemId %d: %v", id, err)
// 				return
// 			}
// 			resultCh <- result{id, res}
// 		}(id)
// 	}

// 	wg.Wait()
// 	close(resultCh)

// 	final := make(map[int][]interface{})
// 	for r := range resultCh {
// 		final[r.id] = r.data
// 	}
// 	return final
// }

func getSchedulesAsync(businessTypeId, businessId int, firstDay, lastDay string, bizItemIdList []int) map[int][]interface{} {
	var wg sync.WaitGroup
	results := make(map[int][]interface{}, len(bizItemIdList))

	var mu sync.Mutex // ✅ map 보호용 뮤텍스

	for _, id := range bizItemIdList {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			res, err := getSchedule(businessTypeId, businessId, firstDay, lastDay, id)
			if err != nil {
				log.Printf("error for bizItemId %d: %v", id, err)
				return
			}
			// for a := range bizItemIdList {
			// 	log.Printf("id: %d, bizItemIdList: %d", id, a)
			// }
			mu.Lock()
			// results[id] = []interface{}{res}
			results[id] = res
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return results
}

func getSchedules(businessTypeId, businessId int, firstDay, lastDay string, bizItemIdList []int) map[int][]interface{} {
	results := make(map[int][]interface{}, len(bizItemIdList))

	for _, id := range bizItemIdList {
		res, err := getSchedule(businessTypeId, businessId, firstDay, lastDay, id)
		if err != nil {
			log.Printf("error for bizItemId %d: %v", id, err)
			continue
		}
		results[id] = res
	}

	return results
}




func getAvailableTimeSlotHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	courtName, ok := request.Params.Arguments["courtName"].(string)
	if !ok {
		return nil, errors.New("failed to get court name")
	}
	yearString, ok := request.Params.Arguments["year"].(string)
	if !ok {
		return nil, errors.New("failed to get year")
	}
	monthString, ok := request.Params.Arguments["month"].(string)
	if !ok {
		return nil, errors.New("failed to get month")
	}

	year, _ := strconv.Atoi(yearString)
	month, _ := strconv.Atoi(monthString)
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month: %d", month)
	}

	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)

	// log.Printf("courtName: %s, year: %d, month: %d", courtName, year, month)

	var businessTypeId = 10
	var businessId int = 0

	maeheonCourtName := []string{"양재", "yangjae", "매헌", "maeheon"}
	lowerCourtName := strings.ToLower(courtName)	
	for _, courtName := range maeheonCourtName {
		if strings.Contains(lowerCourtName, courtName) {
			businessId = 210031
			break
		}
	}

	bizItemIdMap, err := getBizItemIdMap(businessId, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get biz item id map: %v", err)
	}

	bizItemIdList := make([]int, 0, len(bizItemIdMap))
	for k := range bizItemIdMap {
		bizItemIdList = append(bizItemIdList, k)
	}

	// safeCopy := append([]int(nil), bizItemIdList...) // 슬라이스 깊은 복사
	var resultMap = getSchedulesAsync(businessTypeId, businessId, firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"), bizItemIdList)
	// resultList = getSchedulesAsync(businessTypeId, businessId, firstDay.Format("2006-01-02"), lastDay.Format("2006-01-02"), bizItemIdList)

	timetable := make(map[string][]map[string]interface{})
	for bizItemId, times := range resultMap {
		courtNum := bizItemIdMap[bizItemId]
		timetable[courtNum] = make([]map[string]interface{}, 0, len(times))
		for _, playTime := range times {
			ptMap, ok := playTime.(map[string]interface{})
			if !ok {
				continue // 또는 log.Println("unexpected type")
			}

			isUnitBusinessDay, _ := ptMap["isUnitBusinessDay"].(bool)
			unitBookingCount, _ := ptMap["unitBookingCount"].(float64)
			prices, _ := ptMap["prices"].([]interface{}) // JSON 디코딩 결과는 []interface{}일 수 있음
			var price float64
			if len(prices) > 0 {
				if priceMap, ok := prices[0].(map[string]interface{}); ok {
					price, _ = priceMap["price"].(float64)
				}
			}

			unitStartTime, _ := ptMap["unitStartTime"].(string)
			available := isUnitBusinessDay && unitBookingCount == 0

			if (available) {
				timetable[courtNum] = append(timetable[courtNum], map[string]interface{}{
					// "isUnitBusinessDay": isUnitBusinessDay,
					// "unitBookingCount": unitBookingCount,
					"unitStartTime": unitStartTime,
					"available": available,
					"price": price,
				})
			}
		}
	}

	// fmt.Println(resultMap)
	// fmt.Println(bizItemIdList)
	// fmt.Println(bizItemIdMap)

	// key 정렬
	keys := make([]string, 0, len(timetable))
	for k := range timetable {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 정렬된 map을 순서대로 만들어서 JSON 객체처럼 구성
	ordered := make([]byte, 0)
	ordered = append(ordered, '{')

	for i, k := range keys {
		keyBytes, _ := json.Marshal(k)
		valBytes, _ := json.Marshal(timetable[k])

		ordered = append(ordered, keyBytes...)
		ordered = append(ordered, ':')
		ordered = append(ordered, valBytes...)

		if i != len(keys)-1 {
			ordered = append(ordered, ',')
		}
	}

	ordered = append(ordered, '}')

	// jsonTimeTable, _ := json.Marshal(timetable)

	return mcp.NewToolResultText(string(ordered)), nil
}
