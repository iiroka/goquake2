/*
 * Copyright (C) 1997-2001 Id Software, Inc.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or (at
 * your option) any later version.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 *
 * See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307,
 * USA.
 *
 * =======================================================================
 *
 * This file implements the generic part of the menu
 *
 * =======================================================================
 */
package client

const QMF_LEFT_JUSTIFY uint = 0x00000001
const QMF_GRAYED uint = 0x00000002
const QMF_NUMBERSONLY uint = 0x00000004

const RCOLUMN_OFFSET = 16
const LCOLUMN_OFFSET = -16

type menuframework_t struct {
	x, y   int
	cursor int
	owner  *qClient

	nslots int
	items  []menuitem_t

	statusbar string

	// void (*cursordraw)(struct _tag_menuframework *m);
}

type menuitem_t interface {
	setParent(parent *menuframework_t)
	draw()
	getX() int
	getY() int
	getFlags() uint
	getCursorOffset() int
	doEnter() bool
}

type menucommon_t struct {
	name          string
	x, y          int
	parent        *menuframework_t
	cursor_offset int
	// int localdata[4];
	flags uint

	// const char *statusbar;

	callback func(self *menucommon_t)
	// void (*statusbarfunc)(void *self);
	// void (*ownerdraw)(void *self);
	// void (*cursordraw)(void *self);
}

func (T *menucommon_t) setParent(parent *menuframework_t) {
	T.parent = parent
}

func (T *menucommon_t) getX() int {
	return T.x
}

func (T *menucommon_t) getY() int {
	return T.y
}

func (T *menucommon_t) getFlags() uint {
	return T.flags
}

func (T *menucommon_t) getCursorOffset() int {
	return T.cursor_offset
}

type menuaction_t struct {
	menucommon_t
}

func (T *menuframework_t) addItem(item menuitem_t) {
	if T.items == nil {
		T.items = make([]menuitem_t, 0)
		T.nslots = 0
	}

	item.setParent(T)
	T.items = append(T.items, item)

	T.nslots = T.tallySlots()
}

/*
 * This function takes the given menu, the direction, and attempts
 * to adjust the menu's cursor so that it's at the next available
 * slot.
 */
func (T *menuframework_t) adjustCursor(dir int) {
	//  menucommon_s *citem;

	//  /* see if it's in a valid spot */
	if (T.cursor >= 0) && (T.cursor < len(T.items)) {
		// 	 if ((citem = Menu_ItemAtCursor(m)) != 0)
		// 	 {
		// 		 if (citem->type != MTYPE_SEPARATOR)
		// 		 {
		// 			 return;
		// 		 }
		// 	 }
	}

	/* it's not in a valid spot, so crawl in the direction
	indicated until we find a valid spot */
	if dir == 1 {
		// 	 while (1) {
		// 		 citem = Menu_ItemAtCursor(m);

		// 		 if (citem) {
		// 			 if (citem->type != MTYPE_SEPARATOR) {
		// 				 break;
		// 			 }
		// 		 }

		// 		 m->cursor += dir;

		// 		 if (m->cursor >= m->nitems) {
		// 			 m->cursor = 0;
		// 		 }
		// 	 }
	} else {
		// 	 while (1) {
		// 		 citem = Menu_ItemAtCursor(m);

		// 		 if (citem) {
		// 			 if (citem->type != MTYPE_SEPARATOR) {
		// 				 break;
		// 			 }
		// 		 }

		// 		 m->cursor += dir;

		// 		 if (m->cursor < 0) {
		// 			 m->cursor = m->nitems - 1;
		// 		 }
		// 	 }
	}
}

func (T *menuframework_t) center() {
	scale := T.owner.scrGetMenuScale()

	height := T.items[len(T.items)-1].getY()
	height += 10

	T.y = (int(float32(T.owner.viddef.height)/scale) - height) / 2
}

func (T *menuframework_t) draw() {
	// int i;
	// menucommon_s *item;
	scale := T.owner.scrGetMenuScale()

	/* draw contents */
	for i := range T.items {
		T.items[i].draw()
		// switch (((menucommon_s *)menu->items[i])->type) {
		// 	case MTYPE_FIELD:
		// 		Field_Draw((menufield_s *)menu->items[i]);
		// 		break;
		// 	case MTYPE_SLIDER:
		// 		Slider_Draw((menuslider_s *)menu->items[i]);
		// 		break;
		// 	case MTYPE_LIST:
		// 		MenuList_Draw((menulist_s *)menu->items[i]);
		// 		break;
		// 	case MTYPE_SPINCONTROL:
		// 		SpinControl_Draw((menulist_s *)menu->items[i]);
		// 		break;
		// 	case MTYPE_ACTION:
		// 		Action_Draw((menuaction_s *)menu->items[i]);
		// 		break;
		// 	case MTYPE_SEPARATOR:
		// 		Separator_Draw((menuseparator_s *)menu->items[i]);
		// 		break;
		// }
	}

	item := T.itemAtCursor()

	// if (item && item->cursordraw)
	// {
	// 	item->cursordraw(item);
	// }
	// else if (menu->cursordraw)
	// {
	// 	menu->cursordraw(menu);
	// }
	// else if (item && (item->type != MTYPE_FIELD))
	// {
	if (item.getFlags() & QMF_LEFT_JUSTIFY) == 0 {
		T.owner.Draw_CharScaled(T.x+int(float32(int(float32(item.getX())/scale)-24+item.getCursorOffset())*scale),
			int(float32(T.y+item.getY())*scale),
			12+(int(T.owner.common.Sys_Milliseconds()/250)&1), scale)
	} else {
		// 	Draw_CharScaled(menu->x + (item->cursor_offset) * scale,
		// 			(menu->y + item->y) * scale,
		// 			12 + ((int)(Sys_Milliseconds() / 250) & 1), scale);
	}
	// }

	// if (item)
	// {
	// 	if (item->statusbarfunc)
	// 	{
	// 		item->statusbarfunc((void *)item);
	// 	}

	// 	else if (item->statusbar)
	// 	{
	// 		Menu_DrawStatusBar(item->statusbar);
	// 	}

	// 	else
	// 	{
	// 		Menu_DrawStatusBar(menu->statusbar);
	// 	}
	// }
	// else
	// {
	// 	Menu_DrawStatusBar(menu->statusbar);
	// }
}

func (T *menuframework_t) selectItem() bool {
	item := T.itemAtCursor()

	if item != nil {
		return item.doEnter()
		// switch (item->type) {
		// 	case MTYPE_FIELD:
		// 		return Field_DoEnter((menufield_s *)item);
		// 	case MTYPE_ACTION:
		// 		Action_DoEnter((menuaction_s *)item);
		// 		return true;
		// 	case MTYPE_LIST:
		// 		return false;
		// 	case MTYPE_SPINCONTROL:
		// 		return false;
		// }
	}

	return false
}

func (T *menuframework_t) itemAtCursor() menuitem_t {
	if (T.cursor < 0) || (T.cursor >= len(T.items)) {
		return nil
	}

	return T.items[T.cursor]
}

func (T *menuframework_t) tallySlots() int {
	total := 0

	for _ = range T.items {
		// if list, ok := T.items[i].(*menulist_t); ok {
		// if (((menucommon_s *)menu->items[i])->type == MTYPE_LIST) {
		// 	int nitems = 0;
		// 	const char **n = ((menulist_s *)menu->items[i])->itemnames;

		// 	while (*n) {
		// 		nitems++, n++;
		// 	}

		// 	total += nitems;
		// } else {
		total++
		// }
	}

	return total
}

func (T *qClient) menuDrawString(x, y int, str string) {
	scale := T.scrGetMenuScale()

	for i := range str {
		T.Draw_CharScaled(x+int(float32(i*8)*scale), int(float32(y)*scale), int(str[i]), scale)
	}
}

func (T *menuaction_t) draw() {
	Q := T.parent.owner
	scale := Q.scrGetMenuScale()

	if (T.flags & QMF_LEFT_JUSTIFY) != 0 {
		if (T.flags & QMF_GRAYED) != 0 {
			// 		Menu_DrawStringDark(a->generic.x + a->generic.parent->x + (LCOLUMN_OFFSET * scale),
			// 				a->generic.y + a->generic.parent->y, a->generic.name);
		} else {
			Q.menuDrawString(T.x+T.parent.x+int(LCOLUMN_OFFSET*scale),
				T.y+T.parent.y, T.name)
		}
	} else {
		if (T.flags & QMF_GRAYED) != 0 {
			// 		Menu_DrawStringR2LDark(a->generic.x + a->generic.parent->x + (LCOLUMN_OFFSET * scale),
			// 				a->generic.y + a->generic.parent->y, a->generic.name);
		} else {
			// 		Menu_DrawStringR2L(a->generic.x + a->generic.parent->x + (LCOLUMN_OFFSET * scale),
			// 				a->generic.y + a->generic.parent->y, a->generic.name);
		}
	}

	// if (a->generic.ownerdraw) {
	// 	a->generic.ownerdraw(a);
	// }
}
func (T *menuaction_t) doEnter() bool {
	if T.callback != nil {
		T.callback(&T.menucommon_t)
	}
	return true
}