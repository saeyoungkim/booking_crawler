package main

import (
	"booking/data"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

const PAGE_LOAD = 25
const DIALOG_PATH = `div.bui-modal--active div[role="dialog"]`
const DIALOG_CLOSE_PATH = `div[role="dialog"] > div > div > div > div > button`
const LOAD_BUTTON_XPATH = "//span[text()='Load more results']"
const HOTEL_NAME_PATH = "div#hp_hotel_name h2"
const END_LAYOUT_PATH = "div.bottom_of_basiclayout"
const FOOTER_PATH = "#footer_menu_track"
const TAG_PATH = `span[data-testid="facility-icon"] ~ div > span > div > span`
const DESCRIPTION_PATH = `p[data-testid="property-description"]`
const AVAIALABILITY_PATH = `table#hprt-table tbody tr`
const ROOMTYPE_CELL_PATH = "td.hprt-table-cell-roomtype"
const ADULTS_OCCUPANCY_PATH = "span.c-occupancy-icons__adults > i"
const CHILDREN_OCCUPANCY_PATH = "span.c-occupancy-icons__children > i"
const OCCUPANCY_PATH = "td.hprt-table-cell-occupancy span.bui-u-sr-only"
const PRICE_CELL_PATH = "td.hprt-table-cell-price"
const CONDITIONS_CELL_PATH = "td.hprt-table-cell-conditions"
const ROOM_MODAL_OPEN_PATH = `a.hprt-roomtype-link`
const ROOM_MODAL_PATH = `div[role="dialog"] > div[data-component="hotel/new-rooms-table/lightbox"]`
const ROOM_NAME_PATH = "span.hprt-roomtype-icon-link"
const ROOM_DETAIL_CONTAINER = `div.hprt-lightbox-right-container`
const ROOM_SIZE_PATH = `//div[class="hprt-lightbox-right-container"]/h2[1]/following-sibling::text()[1]`
const ROOM_DETAIL_CONTAINER_IF_NO_PICTURE = `div.hprt-lightbox-left-container`
const ROOM_SIZE_PATH_IF_NO_PICTURE = `//div[class="hprt-lightbox-left-container"]/h2[1]/following-sibling::text()[1]`
const ROOM_TYPE_PATH = `div.hprt-roomtype-bed`
const ROOM_TAGS_DIV_PATH = `div.hprt-facilities-facility`
const ROOM_TAGS_SPAN_PATH = `span.hprt-facilities-facility`
const FEE_PATH = `div.prd-taxes-and-fees-under-price`
const CANCELATION_PATH = `ul li.e2e-cancellation`
const ROOM_MODAL_CLOSE_BTN_PATH = "button.modal-mask-closeBtn"
const POLICY_MODAL_OPEN_PATH = `button[data-testid="policy-modal-trigger"]`
const MEAL_DESCRPTION_PATH = `div.bui-group > div:nth-of-type(1) > div.bui-group > div.bui-group__item:nth-of-type(1) > div`
const CANCELATION_SUMMARY_PATH = `div.bui-group > div:nth-of-type(2) > div.bui-group > div.bui-group__item:nth-of-type(1) > div.bui-group:nth-of-type(1) > div.bui-group__item:nth-of-type(2) > div.bui-group:nth-of-type(1) > div.bui-group__item:nth-of-type(1) > span > strong`
const FREE_CANCELATION_STRONG_PATH = "//strong[contains(text(),'Free cancellation')]"
const POLICY_MODAL_CLOSE_BTN_PATH = `button[data-bui-ref="modal-close"]`
const POLICY_MODAL_PATH = `div[data-bui-ref="modal-content-wrapper"]`
const GUEST_REVIEW_PATH = `div[data-testid="PropertyReviewsRegionBlock"]`
const ALL_REVIEW_SCORE_PATH = `div[data-testid="review-score-right-component"] > div:nth-of-type(1)`
const ALL_REVIEW_COUNT_PATH = `div[data-testid="review-score-right-component"] > div:nth-of-type(2) > div:nth-of-type(2)`
const REVIEW_SUBSCORE_PATH = `div[data-testid="PropertyReviewsRegionBlock"] div[data-testid="review-subscore"]`
const SUB_CATEGORY_NAME_AND_SCORE_PATH = `div > div:nth-of-type(1) > div:nth-of-type(1)`
const ROOM_DETAIL_DIALOG_PATH = `div[aria-label="dialog"]`
const HOTEL_ADDRESS_LINK_PATH = `a#hotel_address`

func searchAccomodationLinks(ctx context.Context, destination string, checkin string, checkout string, adults int, children int, rooms int) ([]string, error) {
	indexUrl := fmt.Sprintf("https://www.booking.com/searchresults.en-gb.html?ss=%s&checkin=%s&checkout=%s&group_adults=%d&no_rooms=%d&group_children=%d&nflt=%s&selected_currency=USD&soz=1&lang_changed=1&lang=en-us",
		destination, checkin, checkout, adults, rooms, children, "fc%3D2%3Bht_id%3D204", // HOTEL AND FREE CANCELATION
	)

	// var errOnCloseModal error
	var accommodations []*cdp.Node

	if err := chromedp.Run(ctx,
		chromedp.Navigate(indexUrl),
		chromedp.Sleep(5*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var closeButtonNodes []*cdp.Node
			if err := chromedp.Nodes(DIALOG_CLOSE_PATH, &closeButtonNodes, chromedp.ByQuery, chromedp.AtLeast(0)).Do(ctx); err != nil {
				return err
			}

			if len(closeButtonNodes) > 0 {
				chromedp.Click([]cdp.NodeID{closeButtonNodes[0].NodeID}, chromedp.ByNodeID).Do(ctx)
			}

			return nil
		}),
		chromedp.WaitVisible("body", chromedp.ByQuery),
	); err != nil {
		return nil, err
	}

	prevLen := 0

	log.Println("=====================================")
	log.Println("SCROLLING START")
	log.Println("=====================================")

	// scroll to the end
	for {
		err := chromedp.Run(ctx,
			chromedp.Sleep(5*time.Second),
			chromedp.Nodes(`a[data-testid="availability-cta-btn"]`, &accommodations, chromedp.ByQueryAll),
		)

		if err != nil {
			return nil, err
		}

		if prevLen == len(accommodations) {
			break
		}

		prevLen = len(accommodations)

		lastId := []cdp.NodeID{accommodations[len(accommodations)-1].NodeID}

		if err2 := chromedp.Run(ctx,
			chromedp.ScrollIntoView(lastId, chromedp.ByNodeID),
		); err2 != nil {
			return nil, err2
		}
	}

	log.Println("=====================================")
	log.Println("SCROLLING END")
	log.Println("=====================================")

	log.Println("=====================================")
	log.Println("LOADING BUTTON START")
	log.Println("=====================================")

	var isEnd = false
	for !isEnd {
		err := chromedp.Run(ctx,
			chromedp.ActionFunc(func(ctx context.Context) error {
				var btnNodes []*cdp.Node

				chromedp.Nodes(LOAD_BUTTON_XPATH, &btnNodes, chromedp.BySearch, chromedp.AtLeast(0)).Do(ctx)

				if len(btnNodes) == 0 {
					isEnd = true
				} else {
					chromedp.Click([]cdp.NodeID{btnNodes[0].NodeID}, chromedp.ByNodeID).Do(ctx)
					chromedp.Sleep(5 * time.Second).Do(ctx)
				}

				return nil
			}),
			chromedp.Nodes(`div[aria-label="Property"] a`, &accommodations, chromedp.ByQueryAll),
		)

		if err != nil {
			return nil, err
		}
	}

	log.Println("=====================================")
	log.Println("LOADING BUTTON END")
	log.Println("=====================================")

	var accomodationLinks []string

	for _, accomodation := range accommodations {
		accomodationLinks = append(accomodationLinks, accomodation.AttributeValue("href"))
	}

	return accomodationLinks, nil
}

func makeHeader(hotel_summary *csv.Writer, room_summary *csv.Writer) {
	hotel_summary.Write(
		[]string{
			"hotel_name",
			"latitude",
			"longitude",
			"tags",
			"score",
			"review_count",
			"review_category_score",
		},
	)

	room_summary.Write(
		[]string{
			"hotel_name",
			"room_name",
			"room_type",
			"room_tags",
			"price",
			"free_cancelation",
			"include_breakfast",
		},
	)

	hotel_summary.Flush()
	room_summary.Flush()
}

func makeCategoryReviewsToOneColumn(categoryReviews []data.CategoryReview) string {
	var reviewsToString []string

	for _, categoryReview := range categoryReviews {
		reviewsToString = append(reviewsToString, fmt.Sprintf("%s:%f", categoryReview.Name, categoryReview.Score))
	}

	return strings.Join(reviewsToString, ",")
}

func makeHotelRow(hotel_summary *csv.Writer, name string, latitude string, longitude string, tags []string, score float64, review_count int64, categoryReviews []data.CategoryReview) {
	hotel_summary.Write(
		[]string{
			name,
			latitude,
			longitude,
			strings.Join(tags, ","),
			strconv.FormatFloat(score, 'f', -1, 64),
			strconv.FormatInt(review_count, 10),
			makeCategoryReviewsToOneColumn(categoryReviews),
		},
	)
	hotel_summary.Flush()
}

func makeRoomRow(room_summary *csv.Writer, name string, roomName string, roomType string, roomTags []string, price float64, canCancelFree bool, isIncludedBreakfast bool) {
	room_summary.Write(
		[]string{
			name,
			roomName,
			roomType,
			strings.Join(roomTags, ","),
			strconv.FormatFloat(price, 'f', -1, 64),
			strconv.FormatBool(canCancelFree),
			strconv.FormatBool(isIncludedBreakfast),
		},
	)
	room_summary.Flush()
}

func getInformation(ctx context.Context, accommodationLink string, adults int, children int, hotel_summary *csv.Writer, room_summary *csv.Writer) error {
	var name string

	var latitude string
	var longitude string

	var tags []string
	var description string

	var roomType string = ""
	var roomName string = ""
	var roomTags []string = nil

	var price float64

	var canCancelFree bool
	var isIncludeBreakfast bool

	var allScore float64
	var allReviewes int64

	var categoryReviews []data.CategoryReview

	if err := chromedp.Run(ctx,
		chromedp.Navigate(accommodationLink),
		chromedp.Sleep(8*time.Second),
		// add name
		chromedp.Text(HOTEL_NAME_PATH, &name, chromedp.ByQuery),
		// add address
		chromedp.ActionFunc(func(ctx context.Context) error {
			var addressNodes []*cdp.Node

			if err := chromedp.Nodes(HOTEL_ADDRESS_LINK_PATH, &addressNodes, chromedp.ByQueryAll).Do(ctx); err != nil {
				return err
			}

			var tmp string = addressNodes[0].AttributeValue("data-atlas-latlng")

			var latlng = strings.Split(tmp, ",")

			latitude = latlng[0]
			longitude = latlng[1]

			return nil
		}),
		// add tags
		chromedp.ActionFunc(func(ctx context.Context) error {
			var tagNodes []*cdp.Node
			if err := chromedp.Nodes(TAG_PATH, &tagNodes, chromedp.ByQueryAll).Do(ctx); err != nil {
				return err
			}

			var tag string
			for _, tagNode := range tagNodes {
				var tmp = []cdp.NodeID{tagNode.NodeID}
				if err := chromedp.Text(tmp, &tag, chromedp.ByNodeID).Do(ctx); err != nil {
					return err
				}

				tags = append(tags, tag)
			}

			return nil
		}),
		// add description
		chromedp.Text(DESCRIPTION_PATH, &description, chromedp.ByQuery),
		// export reviews
		chromedp.ActionFunc(func(ctx context.Context) error {
			var scoreStr string

			chromedp.Text(ALL_REVIEW_SCORE_PATH, &scoreStr, chromedp.ByQuery).Do(ctx)

			var scoreParsed = strings.Split(strings.ReplaceAll(scoreStr, "\r\n", "\n"), "\n")[1]

			reviewScore, _ := strconv.ParseFloat(scoreParsed, 64)

			allScore = reviewScore

			var reviewCountStr string

			chromedp.Text(ALL_REVIEW_COUNT_PATH, &reviewCountStr, chromedp.ByQuery).Do(ctx)

			reviewCountStrReplaced := strings.Replace(strings.Split(reviewCountStr, " ")[0], ",", "", 1)

			reviewCount, _ := strconv.ParseInt(reviewCountStrReplaced, 0, 32)

			allReviewes = reviewCount

			var subScoreNodes []*cdp.Node

			chromedp.Nodes(REVIEW_SUBSCORE_PATH, &subScoreNodes, chromedp.ByQueryAll).Do(ctx)

			for _, subScoreNode := range subScoreNodes {
				var subCategoryScoreTmp string

				chromedp.Text(SUB_CATEGORY_NAME_AND_SCORE_PATH, &subCategoryScoreTmp, chromedp.ByQuery, chromedp.FromNode(subScoreNode)).Do(ctx)

				var subCategory = strings.Split(strings.ReplaceAll(subCategoryScoreTmp, "\r\n", "\n"), "\n")

				subCategoryName := subCategory[0]
				subCategoryScore, _ := strconv.ParseFloat(subCategory[1], 64)

				categoryReviews = append(categoryReviews, data.CategoryReview{Name: subCategoryName, Score: subCategoryScore})
			}

			return nil
		}),
		// create Room summary
		chromedp.ActionFunc(func(ctx context.Context) error {
			makeHotelRow(
				hotel_summary,
				name, latitude, longitude, tags, allScore, allReviewes, categoryReviews,
			)

			return nil
		}),
		// export room details
		chromedp.ActionFunc(func(ctx context.Context) error {
			var availabilityNodes []*cdp.Node

			if err := chromedp.Nodes(AVAIALABILITY_PATH, &availabilityNodes, chromedp.ByQueryAll, chromedp.AtLeast(0)).Do(ctx); err != nil {
				return err
			}

			for _, availabilityNode := range availabilityNodes {
				var td []*cdp.Node

				chromedp.Nodes(ROOMTYPE_CELL_PATH, &td, chromedp.ByQuery, chromedp.FromNode(availabilityNode), chromedp.AtLeast(0)).Do(ctx)

				if len(td) > 0 {
					// roomType
					chromedp.Text(ROOM_TYPE_PATH, &roomType, chromedp.ByQuery, chromedp.FromNode(td[0]), chromedp.AtLeast(0)).Do(ctx)
					chromedp.Text(ROOM_NAME_PATH, &roomName, chromedp.ByQuery, chromedp.FromNode(td[0]), chromedp.AtLeast(0)).Do(ctx)
					// chromedp.Text(ROOM_SIZE_PATH, &roomSize, chromedp.BySearch).Do(ctx)

					roomTags = nil

					var roomTagDivNodes []*cdp.Node = nil
					chromedp.Nodes(ROOM_TAGS_DIV_PATH, &roomTagDivNodes, chromedp.ByQueryAll, chromedp.FromNode(td[0]), chromedp.AtLeast(0)).Do(ctx)

					for _, roomTagNode := range roomTagDivNodes {
						tmp := ""
						chromedp.Text([]cdp.NodeID{roomTagNode.NodeID}, &tmp, chromedp.ByNodeID).Do(ctx)

						if len(tmp) > 0 {
							roomTags = append(roomTags, tmp)
						}
					}

					var roomTagSpanNodes []*cdp.Node = nil
					chromedp.Nodes(ROOM_TAGS_SPAN_PATH, &roomTagSpanNodes, chromedp.ByQueryAll, chromedp.FromNode(td[0]), chromedp.AtLeast(0)).Do(ctx)
					for _, roomTagNode := range roomTagSpanNodes {
						tmp := ""
						chromedp.Text([]cdp.NodeID{roomTagNode.NodeID}, &tmp, chromedp.ByNodeID).Do(ctx)

						if len(tmp) > 0 {
							roomTags = append(roomTags, tmp)
						}
					}
				}

				// occupancy
				var adultAndChildren string

				var adultsCount int = 0
				var childrenCount int = 0

				chromedp.Text(OCCUPANCY_PATH, &adultAndChildren, chromedp.ByQuery, chromedp.FromNode(availabilityNode)).Do(ctx)

				occupancies := strings.Split(strings.Trim(adultAndChildren, " "), "<br>")

				if len(occupancies) > 0 {
					adultsCountParsed, _ := strconv.ParseInt(strings.Split(occupancies[0], " ")[2], 0, 32)

					adultsCount = int(adultsCountParsed)
				}

				if len(occupancies) == 2 {
					childrenCountParsed, _ := strconv.ParseInt(strings.Split(occupancies[1], " ")[2], 0, 32)

					childrenCount = int(childrenCountParsed)
				}

				if adultsCount != adults || childrenCount != children {
					continue
				}

				roomPrice, _ := strconv.ParseFloat(availabilityNode.AttributeValue("data-hotel-rounded-price"), 64)

				var roomChrageNodes []*cdp.Node

				chromedp.Nodes(FEE_PATH, &roomChrageNodes, chromedp.ByQueryAll, chromedp.FromNode(availabilityNode)).Do(ctx)

				roomCharge, _ := strconv.ParseFloat(roomChrageNodes[0].AttributeValue("data-excl-charges-raw"), 64)

				price = roomPrice + roomCharge

				conditions := ""

				chromedp.Text(CONDITIONS_CELL_PATH, &conditions, chromedp.ByQuery, chromedp.FromNode(availabilityNode)).Do(ctx)

				isIncludeBreakfast = strings.Contains(conditions, "Breakfast") || strings.Contains(conditions, "breakfast")
				canCancelFree = strings.Contains(conditions, "Free cancellation")

				// // open free cancellation and Meals
				// chromedp.Click(POLICY_MODAL_OPEN_PATH, chromedp.ByQuery, chromedp.FromNode(availabilityNode)).Do(ctx)
				// chromedp.WaitVisible(POLICY_MODAL_PATH, chromedp.ByQuery).Do(ctx)

				// var dialogNodes []*cdp.Node
				// chromedp.Nodes(DIALOG_PATH, &dialogNodes, chromedp.ByQuery).Do(ctx)

				// if len(dialogNodes) > 0 {
				// 	var mealPlan string
				// 	var cancelation string

				// 	chromedp.Text(MEAL_DESCRPTION_PATH, &mealPlan).Do(ctx)
				// 	chromedp.Text(CANCELATION_SUMMARY_PATH, &cancelation).Do(ctx)

				// 	isIncludeBreakfast = !(strings.Contains(mealPlan, "No meal is included") || strings.Contains(mealPlan, "breakfast costs"))
				// 	canCancelFree = strings.Contains(cancelation, "Free cancellation")

				// 	chromedp.Click(POLICY_MODAL_CLOSE_BTN_PATH, chromedp.ByQuery).Do(ctx)
				// 	// chromedp.Sleep(3 * time.Second).Do(ctx)
				// 	chromedp.WaitNotPresent(POLICY_MODAL_PATH, chromedp.ByQuery).Do(ctx)
				// }

				makeRoomRow(
					room_summary,
					name, roomName, roomType, roomTags, price, canCancelFree, isIncludeBreakfast,
				)
			}
			return nil
		}),
	); err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // headless=false に変更
		chromedp.Flag("disable-features", "Translate"),
		chromedp.Flag("disable-notifications", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithDebugf(log.Printf))
	defer cancel()

	adults := 2
	children := 0
	rooms := 1

	city := "Ameritania at Times Square"

	from := "2024-10-10"
	to := "2024-10-11"

	accomodationLinks, err := searchAccomodationLinks(
		ctx,
		city,
		from, to,
		adults, children, rooms,
	)

	if err != nil {
		log.Fatal(err)
	}

	t := time.Now()

	hotel_summary_f, err := os.Create(fmt.Sprintf("%s_%s_%s__%s_hotel_list.csv", city, from, to, t.Format("20060102150405")))

	if err != nil {
		fmt.Println(err)
	}

	room_summary_f, err := os.Create(fmt.Sprintf("%s_%s_%s__%s_room_list.csv", city, from, to, t.Format("20060102150405")))

	if err != nil {
		fmt.Println(err)
	}

	hotel_summary := csv.NewWriter(hotel_summary_f)
	room_summary := csv.NewWriter(room_summary_f)

	makeHeader(hotel_summary, room_summary)

	for _, link := range accomodationLinks {
		linkWithEnglish := link + "&soz=1&lang_changed=1&lang=en-us&selected_currency=USD"
		getInformation(ctx, linkWithEnglish, adults, children, hotel_summary, room_summary)
	}
}
