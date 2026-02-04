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
		Name:    node.Name,
		Home:    node.Home.String(),
		Type:    node.Type.String(),
		Kind:    node.Type.String(),
		Size:    node.Size,
		Weblink: node.Weblink,
		Rev:     node.Rev,
		GRev:    node.GRev,
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

// TrashItemToDTO converts a TrashItem entity to a TrashFolderItem DTO.
func (p *Presenter) TrashItemToDTO(item *entity.TrashItem) TrashFolderItem {
	fi := FolderItem{
		Name: item.Name,
		Home: item.Home.String(),
		Type: item.Type.String(),
		Kind: item.Type.String(),
		Size: item.Size,
		Rev:  item.Rev,
		GRev: item.GRev,
	}

	if item.Type == vo.NodeTypeFile {
		fi.Hash = item.Hash.String()
		fi.MTime = item.MTime
		fi.VirusScan = "pass"
	} else {
		fi.Tree = item.Tree
	}

	return TrashFolderItem{
		FolderItem:  fi,
		DeletedAt:   item.DeletedAt,
		DeletedFrom: item.DeletedFrom,
		DeletedBy:   item.DeletedBy,
	}
}

// ShareToMember converts a Share entity to a ShareMember DTO.
func (p *Presenter) ShareToMember(share *entity.Share) ShareMember {
	return ShareMember{
		Email:  share.InvitedEmail,
		Status: share.Status.String(),
		Access: share.Access.String(),
		Name:   share.InvitedEmail,
	}
}

// ShareToIncomingInvite converts a Share entity to an IncomingInvite DTO.
// ownerEmail is the email of the share owner, resolved by the caller.
func (p *Presenter) ShareToIncomingInvite(share *entity.Share, ownerEmail string) IncomingInvite {
	return IncomingInvite{
		Owner: InviteOwner{
			Email: ownerEmail,
			Name:  ownerEmail,
		},
		Access:      share.Access.String(),
		Name:        share.Home.Name(),
		Home:        share.Home.String(),
		InviteToken: share.InviteToken,
	}
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
