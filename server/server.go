package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	// "io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/braintree/manners"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/jasonlvhit/gocron"
	"github.com/spf13/cobra"

	"github.com/cpanato/mattermost-away-reminder/model"
	"github.com/cpanato/mattermost-away-reminder/store"
)

type Server struct {
	Store  store.Store
	Router *mux.Router
}

const (
	LAYOUT = "2006/01/02"
)

var (
	Srv *Server
)

func Start() {
	fmt.Println("Starting Away Bot")

	Srv = &Server{
		Store:  store.NewSqlStore(Config.DriverName, Config.DataSource),
		Router: mux.NewRouter(),
	}

	addApis(Srv.Router)

	var handler http.Handler = Srv.Router
	go func() {
		fmt.Println("Listening on %v", Config.ListenAddress)
		err := manners.ListenAndServe(Config.ListenAddress, handler)
		if err != nil {
			fmt.Println(err.Error())
		}

	}()

	gocron.Every(12).Hours().Do(postAways)
	gocron.Every(1).Day().Do(removeOldAways)
	fmt.Println("Starting timer")
	<-gocron.Start()
}

func addApis(r *mux.Router) {
	r.HandleFunc("/", indexHandler).Methods("GET", "POST")
	r.HandleFunc("/ping", pingHandler).Methods("GET")
	// r.HandleFunc("/delete", deleteUserAways).Methods("POST")
	r.HandleFunc("/time_away", slashCommandHandler).Methods("POST")
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("This is the mattermost Time Away Bot Server."))
}

func pingHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("ok"))
}

// func deleteUserAways(w http.ResponseWriter, r *http.Request) {
// 	resp := "Deleted"
// 	body, _ := ioutil.ReadAll(r.Body)
// 	msg := string(body)

// 	WriteIntegrationResponse(w, resp)
// }

func ParseSlashCommand(r *http.Request) (*model.MMSlashCommand, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	inCommand := &model.MMSlashCommand{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	err = decoder.Decode(inCommand, r.Form)
	if err != nil {
		return nil, err
	}

	return inCommand, nil
}

func checkSlashPermissions(command *model.MMSlashCommand) *model.AppError {
	hasPremissions := false
	for _, allowedToken := range Config.AllowedTokens {
		if allowedToken == command.Token {
			hasPremissions = true
			break
		}
	}

	if !hasPremissions {
		return model.NewLocAppError("Mattermost Time Away", "Token for slash command is incorrect", nil, "")
	}

	return nil
}

func slashCommandHandler(w http.ResponseWriter, r *http.Request) {
	command, err := ParseSlashCommand(r)
	if err != nil {
		WriteErrorResponse(w, model.NewLocAppError("Mattermost Time Away", "Unable to parse incoming slash command info", nil,
			err.Error()))
		return
	}

	if err := checkSlashPermissions(command); err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// Output Buffer
	outBuf := &bytes.Buffer{}

	var rootCmd = &cobra.Command{
		Use:   "/time_away",
		Short: "Share you time away with your colleagues",
	}

	var saveAwayCmd = &cobra.Command{
		Use:   "save --from [date YYYY/MM/DD] --to [date YYYY/MM/DD] [reason]",
		Short: "Save a time away",
		Long:  "Save a time away",
		RunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")
			return saveAwayCommandF(args, w, command, from, to)
		},
	}
	saveAwayCmd.Flags().StringP("from", "f", time.Now().Format(LAYOUT), "Away Start date")
	saveAwayCmd.Flags().StringP("to", "t", time.Now().Format(LAYOUT), "Away End date")

	var ListUserAwaysCmd = &cobra.Command{
		Use:   "list",
		Short: "List your aways",
		Long:  "List your aways",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listUserAwaysCommandF(args, w, command)
		},
	}

	var ListAllAwaysCmd = &cobra.Command{
		Use:   "listall",
		Short: "List all aways for today",
		Long:  "List all aways for today",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listAllAwaysCommandF(args, w, command)
		},
	}

	var DeleteUserAwaysCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove your aways",
		Long:  "Remove your aways",
		RunE: func(cmd *cobra.Command, args []string) error {
			id, _ := cmd.Flags().GetString("id")
			return deleteUserAwaysCommandF(args, w, command, id)
		},
	}
	DeleteUserAwaysCmd.Flags().StringP("id", "i", "", "Away id to remove")

	rootCmd.SetArgs(strings.Fields(strings.TrimSpace(command.Text)))
	rootCmd.SetOutput(outBuf)

	rootCmd.AddCommand(saveAwayCmd, ListUserAwaysCmd, ListAllAwaysCmd, DeleteUserAwaysCmd)

	err = rootCmd.Execute()

	if err != nil || len(outBuf.String()) > 0 {
		WriteResponse(w, outBuf.String(), "ephemeral")
	}
	return
}

func saveAwayCommandF(args []string, w http.ResponseWriter, slashCommand *model.MMSlashCommand, from, to string) error {
	userName := slashCommand.Username
	if len(args) == 0 {
		return model.NewLocAppError("Mattermost Time Away", "You need to specify a reason", nil, "")
	}

	fromParsed, err := time.Parse(LAYOUT, from)
	if err != nil {
		fmt.Println(err)
		return model.NewLocAppError("Mattermost Time Away", "Error to parse the From date, please use the following format: YYYY/MM/DD", nil, err.Error())
	}

	toParsed, err := time.Parse(LAYOUT, to)
	if err != nil {
		fmt.Println(err)
		return model.NewLocAppError("Mattermost Time Away", "Error to parse the To date, please use the following format: YYYY/MM/DD", nil, err.Error())
	}

	if toParsed.Unix() < fromParsed.Unix() {
		err := model.NewLocAppError("Server.SaveAway", "**To** Date is bigger them **From** date", nil,
			fmt.Sprintf("From=%v, To=%v", from, to))
		return err
	}

	saveAway := &model.Away{
		Start:    strconv.FormatInt(fromParsed.Unix(), 10),
		End:      strconv.FormatInt(toParsed.Unix(), 10),
		UserName: userName,
		Reason:   strings.Join(args, " "),
	}
	if result := <-Srv.Store.Away().Save(saveAway); result.Err != nil {
		return model.NewLocAppError("Mattermost Time Away", "Error to save the time away", nil, result.Err.Error())
	}

	WriteResponse(w, "Away successfully saved", "ephemeral")
	return nil
}

func listUserAwaysCommandF(args []string, w http.ResponseWriter, slashCommand *model.MMSlashCommand) error {
	var aways []*model.Away
	if result := <-Srv.Store.Away().ListUserAway(slashCommand.Username); result.Err != nil {
		fmt.Println(result.Err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return result.Err
	} else {
		aways = result.Data.([]*model.Away)
	}

	attachment1 := []model.MMAttachment{}
	for _, away := range aways {
		startInt, _ := strconv.ParseInt(away.Start, 10, 64)
		endInt, _ := strconv.ParseInt(away.End, 10, 64)

		msg := fmt.Sprintf("Id: %v From: %v To: %v Reason: %v", away.Id, time.Unix(startInt, 0).Format(LAYOUT), time.Unix(endInt, 0).Format(LAYOUT), away.Reason)

		attach := model.MMAttachment{}
		// attachment1 = append(attachment1, *attach.AddField(model.MMField{Title: "Time Away", Value: msg}).AddAction(model.MMAction{Name: "Delete", Integration: &model.MMActionIntegration{URL: "http://localhost:8087/delete", Context: model.StringInterface{"id": away.Id}}}))
		attachment1 = append(attachment1, *attach.AddField(model.MMField{Title: "Time Away", Value: msg}))
	}

	payload := model.MMSlashResponse{
		ResponseType: "ephemeral",
		Text:         "List of Aways for user: " + slashCommand.Username,
		Username:     "TimeAway",
		IconUrl:      "https://png.icons8.com/ios/1600/sunbathe.png",
		Attachments:  attachment1,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Unable to marshal response")
		return model.NewLocAppError("Mattermost Time Away", "Unable to marshal response", nil, "")
	}

	fmt.Println(payload.ToJson())
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return nil
}

func listAllAwaysCommandF(args []string, w http.ResponseWriter, slashCommand *model.MMSlashCommand) error {
	t, err := time.Parse(LAYOUT, time.Now().Format(LAYOUT))
	if err != nil {
		fmt.Println(err)
	}

	var aways []*model.Away
	if result := <-Srv.Store.Away().GetAwaysForToday(strconv.FormatInt(t.Unix(), 10)); result.Err != nil {
		fmt.Println(result.Err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return result.Err
	} else {
		aways = result.Data.([]*model.Away)
	}

	attachment1 := []model.MMAttachment{}
	for _, away := range aways {
		startInt, _ := strconv.ParseInt(away.Start, 10, 64)
		endInt, _ := strconv.ParseInt(away.End, 10, 64)

		msg := fmt.Sprintf("User: @%v From: %v To: %v Reason: %v", away.UserName, time.Unix(startInt, 0).Format(LAYOUT), time.Unix(endInt, 0).Format(LAYOUT), away.Reason)

		attach := model.MMAttachment{}
		// attachment1 = append(attachment1, *attach.AddField(model.MMField{Title: "Time Away", Value: msg}).AddAction(model.MMAction{Name: "Delete", Integration: &model.MMActionIntegration{URL: "http://localhost:8087/delete", Context: model.StringInterface{"id": away.Id}}}))
		attachment1 = append(attachment1, *attach.AddField(model.MMField{Title: "Time Away", Value: msg}))
	}

	payload := model.MMSlashResponse{
		ResponseType: "ephemeral",
		Text:         "List of Aways for all users",
		Username:     "TimeAway",
		IconUrl:      "https://png.icons8.com/ios/1600/sunbathe.png",
		Attachments:  attachment1,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Unable to marshal response")
		return model.NewLocAppError("Mattermost Time Away", "Unable to marshal response", nil, "")
	}

	fmt.Println(payload.ToJson())
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return nil
}

func deleteUserAwaysCommandF(args []string, w http.ResponseWriter, slashCommand *model.MMSlashCommand, id string) error {
	if id == "" {
		return model.NewLocAppError("Mattermost Time Away", "Please use --id and set the id", nil, "")
	}

	result := <-Srv.Store.Away().GetAwaysById(id)
	if result.Err != nil {
		fmt.Println(result.Err.Error())
		return result.Err
	}

	if result.Data.(*model.Away).UserName != slashCommand.Username {
		fmt.Println("Away does not belong to you.")
		return model.NewLocAppError("Mattermost Time Away", "This Away does not belong to you.", nil, "")
	} else {
		if result := <-Srv.Store.Away().DeleteAway(id); result.Err != nil {
			fmt.Println(result.Err.Error())
			return result.Err
		}
	}

	WriteResponse(w, "Id="+id+" successfully deleted.", "ephemeral")
	return nil
}

func WriteErrorResponse(w http.ResponseWriter, err *model.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(model.GenerateStandardSlashResponse(err.Error(), "ephemeral")))
}

func WriteResponse(w http.ResponseWriter, resp string, msgType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(model.GenerateStandardSlashResponse(resp, msgType)))
}

// func WriteIntegrationResponse(w http.ResponseWriter, resp string) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(model.GenerateIntegrationResponse(resp)))
// }
