package clusters

import (
	"net/http"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/lob/pharos/pkg/pharos-api-server/application"
	"github.com/lob/pharos/pkg/util/model"
	"github.com/pkg/errors"
)

type handler struct {
	app application.App
}

type listQuery struct {
	Environment string `query:"environment"`
	Active      bool   `query:"active"`
}

func (h *handler) list(c echo.Context) error {
	clusters := make([]*model.Cluster, 0)

	query := listQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	q := h.app.DB.
		Model(&clusters).
		Where("deleted = FALSE").
		Order("date_created DESC")

	if query.Environment != "" {
		q = q.Where("environment = ?", query.Environment)
	}

	if query.Active {
		q = q.Where("active = ?", query.Active)
	}

	err := q.Select()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, clusters)
}

func (h *handler) retrieve(c echo.Context) error {
	id := c.Param("id")

	var cluster model.Cluster

	err := h.app.DB.Model(&cluster).Where("id = ?", id).First()
	if err != nil {
		if err == pg.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "cluster not found")
		}
		return err
	}

	return c.JSON(http.StatusOK, cluster)
}

func (h *handler) delete(c echo.Context) error {
	id := c.Param("id")

	var cluster model.Cluster

	err := h.app.DB.Model(&cluster).Where("id = ?", id).First()
	if err != nil {
		if err == pg.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "cluster not found")
		}
		return err
	}

	cluster.Deleted = true

	_, err = h.app.DB.Model(&cluster).WherePK().Update()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, cluster)
}

type createParams struct {
	ID                   string `json:"id"                     mod:"trim" validate:"required"`
	Environment          string `json:"environment"            mod:"trim" validate:"required"`
	ServerURL            string `json:"server_url"             mod:"trim" validate:"required,url"`
	ClusterAuthorityData string `json:"cluster_authority_data" mod:"trim" validate:"required,base64"`
}

func (h *handler) create(c echo.Context) error {
	params := createParams{}
	if err := c.Bind(&params); err != nil {
		return err
	}

	cluster := model.Cluster{
		ID:                   params.ID,
		Environment:          params.Environment,
		ServerURL:            params.ServerURL,
		ClusterAuthorityData: params.ClusterAuthorityData,
	}

	_, err := h.app.DB.Model(&cluster).Insert()
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, cluster)
}

type updateParams struct {
	Active bool `json:"active" validate:"required"`
}

func (h *handler) update(c echo.Context) error {
	id := c.Param("id")

	params := updateParams{}
	if err := c.Bind(&params); err != nil {
		return err
	}

	var cluster model.Cluster

	err := h.app.DB.Model(&cluster).Where("id = ?", id).Where("deleted = FALSE", id).First()
	if err != nil {
		if err == pg.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "cluster not found")
		}
		return err
	}

	cluster.Active = params.Active

	err = h.app.DB.RunInTransaction(func(tx *pg.Tx) error {
		if _, err := tx.Model(&model.Cluster{}).Set("active = FALSE").Where("environment = ?", cluster.Environment).Update(); err != nil {
			return err
		}

		_, err := tx.Model(&cluster).WherePK().Update()
		return err
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return c.JSON(http.StatusOK, cluster)
}
