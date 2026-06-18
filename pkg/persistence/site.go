// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package persistence

import (
	"context"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"gorm.io/gorm"
)

// SiteRepository manages physical storefronts/facilities and sub-locations (fixtures/shelves).
type SiteRepository interface {
	Create(ctx context.Context, s *model.Site) error
	FindByID(ctx context.Context, id string) (*model.Site, error)
	Update(ctx context.Context, s *model.Site) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*model.Site, error)
	ListRange(ctx context.Context, offset, limit int) ([]*model.Site, error)
	CreateLocation(ctx context.Context, l *model.Location) error
	FindLocationByID(ctx context.Context, id string) (*model.Location, error)
	UpdateLocation(ctx context.Context, l *model.Location) error
	DeleteLocation(ctx context.Context, id string) error
	ListLocations(ctx context.Context) ([]*model.Location, error)
	ListLocationsRange(ctx context.Context, offset, limit int) ([]*model.Location, error)
	CreateAsset(ctx context.Context, a *model.Asset) error
	FindAssetByID(ctx context.Context, id string) (*model.Asset, error)
	UpdateAsset(ctx context.Context, a *model.Asset) error
	DeleteAsset(ctx context.Context, id string) error
	ListAssets(ctx context.Context) ([]*model.Asset, error)
	ListAssetsRange(ctx context.Context, offset, limit int) ([]*model.Asset, error)
}

type siteRepository struct {
	db *gorm.DB
}

// NewSiteRepository creates a new SiteRepository instance.
func NewSiteRepository(db *gorm.DB) SiteRepository {
	return &siteRepository{db: db}
}

func (r *siteRepository) Create(ctx context.Context, s *model.Site) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *siteRepository) FindByID(ctx context.Context, id string) (*model.Site, error) {
	var s model.Site
	err := r.db.WithContext(ctx).Preload("Locations").First(&s, "id = ?", id).Error
	return &s, err
}

func (r *siteRepository) Update(ctx context.Context, s *model.Site) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *siteRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Site{}, "id = ?", id).Error
}

func (r *siteRepository) List(ctx context.Context) ([]*model.Site, error) {
	var list []*model.Site
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *siteRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.Site, error) {
	var list []*model.Site
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *siteRepository) CreateLocation(ctx context.Context, l *model.Location) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *siteRepository) FindLocationByID(ctx context.Context, id string) (*model.Location, error) {
	var l model.Location
	err := r.db.WithContext(ctx).First(&l, "id = ?", id).Error
	return &l, err
}

func (r *siteRepository) UpdateLocation(ctx context.Context, l *model.Location) error {
	return r.db.WithContext(ctx).Save(l).Error
}

func (r *siteRepository) DeleteLocation(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Location{}, "id = ?", id).Error
}

func (r *siteRepository) ListLocations(ctx context.Context) ([]*model.Location, error) {
	var list []*model.Location
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *siteRepository) ListLocationsRange(ctx context.Context, offset, limit int) ([]*model.Location, error) {
	var list []*model.Location
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}

func (r *siteRepository) CreateAsset(ctx context.Context, a *model.Asset) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *siteRepository) FindAssetByID(ctx context.Context, id string) (*model.Asset, error) {
	var a model.Asset
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	return &a, err
}

func (r *siteRepository) UpdateAsset(ctx context.Context, a *model.Asset) error {
	return r.db.WithContext(ctx).Save(a).Error
}

func (r *siteRepository) DeleteAsset(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Asset{}, "id = ?", id).Error
}

func (r *siteRepository) ListAssets(ctx context.Context) ([]*model.Asset, error) {
	var list []*model.Asset
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *siteRepository) ListAssetsRange(ctx context.Context, offset, limit int) ([]*model.Asset, error) {
	var list []*model.Asset
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}
