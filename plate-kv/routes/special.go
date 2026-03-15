// Endpoints contained in this file:
// POST /{plateID}/bitmaps/command
// POST /{plateID}/bitmaps/{key}/command
// POST /{plateID}/bitmaps/set
// GET /{plateID}/bitmaps/get/{key}/{bit}
// GET /{plateID}/bitmaps/count/{key}
// GET /{plateID}/bitmaps/position/{key}/{bit}
// POST /{plateID}/bitmaps/operate
// POST /{plateID}/bitmaps/field
// POST /{plateID}/geo/command
// POST /{plateID}/geo/{key}/command
// POST /{plateID}/geo/add
// POST /{plateID}/geo/positions
// GET /{plateID}/geo/distance/{key}
// POST /{plateID}/geo/search
// POST /{plateID}/geo/search/store
package routes

import (
	"net/http"
	"strings"

	"plain/kv/internal/plate"
)

func registerSpecial(mux *http.ServeMux, deps *plate.Dependencies) {
	bitmapAllowed := mustCommands("SETBIT", "GETBIT", "BITCOUNT", "BITOP", "BITPOS", "BITFIELD")
	geoAllowed := mustCommands("GEOADD", "GEOPOS", "GEODIST", "GEOSEARCH", "GEOSEARCHSTORE")
	mux.HandleFunc("POST /{plateID}/bitmaps/command", handleCommand(deps, bitmapAllowed, false))
	mux.HandleFunc("POST /{plateID}/bitmaps/{key}/command", handleCommand(deps, bitmapAllowed, true))
	mux.HandleFunc("POST /{plateID}/geo/command", handleCommand(deps, geoAllowed, false))
	mux.HandleFunc("POST /{plateID}/geo/{key}/command", handleCommand(deps, geoAllowed, true))
	mux.HandleFunc("POST /{plateID}/bitmaps/set", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key   string `json:"key"`
			Bit   int64  `json:"bit"`
			Value int    `json:"value"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if request.Bit < 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "bit must be zero or greater")
		}
		if request.Value != 0 && request.Value != 1 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "value must be 0 or 1")
		}
		result, err := execute(r, deps, plateID, "SETBIT", key, request.Bit, request.Value)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/bitmaps/get/{key}/{bit}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		bit, err := plate.PathValue(r, "bit")
		if err != nil {
			return err
		}
		result, err := execute(r, deps, plateID, "GETBIT", key, bit)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/bitmaps/count/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		args := []any{key}
		if start := strings.TrimSpace(r.URL.Query().Get("start")); start != "" {
			args = append(args, start)
			args = append(args, queryString(r, "end", "-1"))
		}
		result, err := execute(r, deps, plateID, "BITCOUNT", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/bitmaps/position/{key}/{bit}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		bit, err := plate.PathValue(r, "bit")
		if err != nil {
			return err
		}
		args := []any{key, bit}
		if start := strings.TrimSpace(r.URL.Query().Get("start")); start != "" {
			args = append(args, start)
			if end := strings.TrimSpace(r.URL.Query().Get("end")); end != "" {
				args = append(args, end)
			}
		}
		result, err := execute(r, deps, plateID, "BITPOS", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/bitmaps/operate", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Operation   string   `json:"operation"`
			Destination string   `json:"destination"`
			Sources     []string `json:"sources"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		operation := strings.ToUpper(strings.TrimSpace(request.Operation))
		if operation != "AND" && operation != "OR" && operation != "XOR" && operation != "NOT" {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "operation must be and, or, xor, or not")
		}
		destination, err := requiredString(request.Destination, "destination")
		if err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.Sources, "sources"); err != nil {
			return err
		}
		if operation == "NOT" && len(request.Sources) != 1 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "not requires exactly one source")
		}
		args := []any{operation, destination}
		for _, source := range request.Sources {
			args = append(args, source)
		}
		result, err := execute(r, deps, plateID, "BITOP", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/bitmaps/field", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key        string `json:"key"`
			Operations []any  `json:"operations"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if len(request.Operations) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "operations must not be empty")
		}
		args := append([]any{key}, request.Operations...)
		result, err := execute(r, deps, plateID, "BITFIELD", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/geo/add", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key       string `json:"key"`
			Locations []struct {
				Member    string `json:"member"`
				Longitude any    `json:"longitude"`
				Latitude  any    `json:"latitude"`
			} `json:"locations"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if len(request.Locations) == 0 {
			return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "locations must not be empty")
		}
		args := []any{key}
		for _, location := range request.Locations {
			member, err := requiredString(location.Member, "member")
			if err != nil {
				return err
			}
			longitude, err := redisValueString(location.Longitude)
			if err != nil {
				return err
			}
			latitude, err := redisValueString(location.Latitude)
			if err != nil {
				return err
			}
			args = append(args, longitude, latitude, member)
		}
		result, err := execute(r, deps, plateID, "GEOADD", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/geo/positions", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key     string   `json:"key"`
			Members []string `json:"members"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		if err := requireNonEmptyStrings(request.Members, "members"); err != nil {
			return err
		}
		args := []any{key}
		for _, member := range request.Members {
			args = append(args, member)
		}
		result, err := execute(r, deps, plateID, "GEOPOS", args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("GET /{plateID}/geo/distance/{key}", plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		key, err := plate.PathValue(r, "key")
		if err != nil {
			return err
		}
		from, err := queryRequiredString(r, "from")
		if err != nil {
			return err
		}
		to, err := queryRequiredString(r, "to")
		if err != nil {
			return err
		}
		unit := queryString(r, "unit", "m")
		result, err := execute(r, deps, plateID, "GEODIST", key, from, to, unit)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	}))
	mux.HandleFunc("POST /{plateID}/geo/search", geoSearchHandler(deps, "GEOSEARCH"))
	mux.HandleFunc("POST /{plateID}/geo/search/store", geoSearchHandler(deps, "GEOSEARCHSTORE"))
}

func geoSearchHandler(deps *plate.Dependencies, command string) http.HandlerFunc {
	return plate.Authenticated(deps, func(w http.ResponseWriter, r *http.Request, plateID string) error {
		var request struct {
			Key         string `json:"key"`
			Destination string `json:"destination"`
			FromMember  string `json:"from_member"`
			FromLon     any    `json:"from_lon"`
			FromLat     any    `json:"from_lat"`
			Radius      any    `json:"radius"`
			Width       any    `json:"width"`
			Height      any    `json:"height"`
			Unit        string `json:"unit"`
			Count       *int64 `json:"count"`
			Any         bool   `json:"any"`
			Sort        string `json:"sort"`
			WithCoord   bool   `json:"with_coord"`
			WithDist    bool   `json:"with_dist"`
			WithHash    bool   `json:"with_hash"`
		}
		if err := plate.DecodeJSON(r, &request); err != nil {
			return err
		}
		key, err := requiredString(request.Key, "key")
		if err != nil {
			return err
		}
		args := []any{key}
		if command == "GEOSEARCHSTORE" {
			destination, err := requiredString(request.Destination, "destination")
			if err != nil {
				return err
			}
			args = append([]any{destination, key}, args[1:]...)
		}
		if member := strings.TrimSpace(request.FromMember); member != "" {
			args = append(args, "FROMMEMBER", member)
		} else {
			lon, err := redisValueString(request.FromLon)
			if err != nil {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "from_member or from_lon/from_lat is required")
			}
			lat, err := redisValueString(request.FromLat)
			if err != nil {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "from_member or from_lon/from_lat is required")
			}
			args = append(args, "FROMLONLAT", lon, lat)
		}
		unit := strings.ToLower(strings.TrimSpace(request.Unit))
		if unit == "" {
			unit = "m"
		}
		if request.Radius != nil {
			radius, err := redisValueString(request.Radius)
			if err != nil {
				return err
			}
			args = append(args, "BYRADIUS", radius, unit)
		} else {
			width, err := redisValueString(request.Width)
			if err != nil {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "radius or width/height is required")
			}
			height, err := redisValueString(request.Height)
			if err != nil {
				return plate.NewAPIError(http.StatusBadRequest, "invalid_request", "radius or width/height is required")
			}
			args = append(args, "BYBOX", width, height, unit)
		}
		sort := strings.ToUpper(strings.TrimSpace(request.Sort))
		if sort == "ASC" || sort == "DESC" {
			args = append(args, sort)
		}
		if request.Count != nil {
			args = append(args, "COUNT", *request.Count)
			if request.Any {
				args = append(args, "ANY")
			}
		}
		if command == "GEOSEARCH" {
			if request.WithCoord {
				args = append(args, "WITHCOORD")
			}
			if request.WithDist {
				args = append(args, "WITHDIST")
			}
			if request.WithHash {
				args = append(args, "WITHHASH")
			}
		}
		result, err := execute(r, deps, plateID, command, args...)
		if err != nil {
			return err
		}
		writeResult(w, result)
		return nil
	})
}
