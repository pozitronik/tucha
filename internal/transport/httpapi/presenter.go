package httpapi

import (
	"tucha/internal/domain/entity"
	"tucha/internal/domain/vo"
)

// Presenter maps domain entities to transport DTOs.
type Presenter struct{}

// NewPresenter creates a new Presenter.
func NewPresenter() *Presenter {
	return &Presenter{}
}

// NodeToFolderItem converts a Node entity to a FolderItem DTO.
// For folders, subCount must be provided (use nil for files).
func (p *Presenter) NodeToFolderItem(node *entity.Node, subCount *FolderCount) FolderItem {
	item := FolderItem{
		Name: node.Name,
		Home: node.Home.String(),
		Type: node.Type.String(),
		Kind: node.Type.String(),
		Size: node.Size,
		Rev:  node.Rev,
		GRev: node.GRev,
	}

	if node.Type == vo.NodeTypeFile {
		item.Hash = node.Hash.String()
		item.MTime = node.MTime
		item.VirusScan = "pass"
	} else {
		item.Tree = node.Tree
		item.Count = subCount
	}

	return item
}

// BuildFolderListing builds a complete FolderListing DTO from a folder node and its children items.
func (p *Presenter) BuildFolderListing(
	folder *entity.Node,
	count FolderCount,
	items []FolderItem,
) FolderListing {
	return FolderListing{
		Count: count,
		Tree:  folder.Tree,
		Name:  folder.Name,
		GRev:  folder.GRev,
		Size:  folder.Size,
		Sort:  SortInfo{Order: "asc", Type: "name"},
		Kind:  "folder",
		Rev:   folder.Rev,
		Type:  "folder",
		Home:  folder.Home.String(),
		List:  items,
	}
}
