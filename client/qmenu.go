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

import "strings"

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
	DoSlide(dir int)
	isField() bool
	listItems() int
}

type menucommon_t struct {
	name          string
	x, y          int
	parent        *menuframework_t
	cursor_offset int
	// int localdata[4];
	flags uint

	statusbar string

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

func (T *menucommon_t) isField() bool {
	return false
}

func (T *menucommon_t) listItems() int {
	return 0
}

func (T *menucommon_t) doEnter() bool {
	return false
}

func (T *menucommon_t) DoSlide(dir int) {}

type menufield_t struct {
	menucommon_t

	buffer         string
	cursor         int
	length         int
	visible_length int
	visible_offset int
}

func (T *menufield_t) isField() bool {
	return true
}

type menuaction_t struct {
	menucommon_t
}

type menulist_t struct {
	menucommon_t
	curvalue  int
	itemnames []string
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

func (T *menuframework_t) SlideItem(dir int) {
	item := T.itemAtCursor()
	if item != nil {
		item.DoSlide(dir)
	}
}

func (T *menuframework_t) center() {
	scale := T.owner.scrGetMenuScale()

	height := T.items[len(T.items)-1].getY()
	height += 10

	T.y = (int(float32(T.owner.viddef.height)/scale) - height) / 2
}

func (T *menuframework_t) setStatusBar(str string) {
	T.statusbar = str
}

func (T *menuframework_t) draw() {
	scale := T.owner.scrGetMenuScale()

	/* draw contents */
	for i := range T.items {
		T.items[i].draw()
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
	// else
	if item != nil && !item.isField() {
		if (item.getFlags() & QMF_LEFT_JUSTIFY) == 0 {
			T.owner.Draw_CharScaled(T.x+int(float32(int(float32(item.getX())/scale)-24+item.getCursorOffset())*scale),
				int(float32(T.y+item.getY())*scale),
				12+(int(T.owner.common.Sys_Milliseconds()/250)&1), scale)
		} else {
			T.owner.Draw_CharScaled(T.x+int(float32(item.getCursorOffset())*scale),
				int(float32(T.y+item.getY())*scale),
				12+(int(T.owner.common.Sys_Milliseconds()/250)&1), scale)
		}
	}

	if item != nil {
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
	} else {
		// 	Menu_DrawStatusBar(menu->statusbar);
	}
}

func (T *menuframework_t) selectItem() bool {
	item := T.itemAtCursor()

	if item != nil {
		return item.doEnter()
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

	for _, item := range T.items {
		nitems := item.listItems()
		if nitems > 0 {
			total += nitems
		} else {
			total++
		}
	}

	return total
}

func (T *menufield_t) doEnter() bool {
	if T.callback != nil {
		T.callback(&T.menucommon_t)
		return true
	}

	return false
}

func (T *menufield_t) draw() {
	Q := T.parent.owner
	scale := Q.scrGetMenuScale()

	if len(T.name) > 0 {
		Q.menuDrawStringR2LDark(T.x+T.parent.x+int(LCOLUMN_OFFSET*scale), T.y+T.parent.y,
			T.name)
	}

	Q.Draw_CharScaled(int(float32(T.x+T.parent.x+16)*scale), int(float32(T.y+T.parent.y-4)*scale), 18, scale)
	Q.Draw_CharScaled(int(float32(T.x+T.parent.x+16)*scale), int(float32(T.y+T.parent.y+4)*scale), 24, scale)

	Q.Draw_CharScaled(int(float32(T.x+T.parent.x+24)*scale)+int(float32(T.visible_length)*8*scale),
		int(float32(T.y+T.parent.y-4)*scale), 20, scale)
	Q.Draw_CharScaled(int(float32(T.x+T.parent.x+24)*scale)+int(float32(T.visible_length)*8*scale),
		int(float32(T.y+T.parent.y+4)*scale), 26, scale)

	for i := 0; i < T.visible_length; i++ {
		Q.Draw_CharScaled(int(float32(T.x+T.parent.x+24)*scale)+int(float32(i)*8*scale),
			int(float32(T.y+T.parent.y-4)*scale), 19, scale)
		Q.Draw_CharScaled(int(float32(T.x+T.parent.x+24)*scale)+int(float32(i)*8*scale),
			int(float32(T.y+T.parent.y+4)*scale), 25, scale)
	}

	n := T.visible_length + 1 + T.visible_offset
	var tempbuffer string
	if n >= len(T.buffer) {
		tempbuffer = T.buffer[T.visible_offset:]
	} else {
		tempbuffer = T.buffer[T.visible_offset:n]
	}
	Q.menuDrawString(int(float32(T.x+T.parent.x+24)*scale),
		T.y+T.parent.y, tempbuffer)

	if T.parent.itemAtCursor() == T {
		offset := 0

		if T.visible_offset != 0 {
			offset = T.visible_length
		} else {
			offset = T.cursor
		}

		if ((Q.common.Sys_Milliseconds() / 250) & 1) != 0 {
			Q.Draw_CharScaled(
				int(float32(T.x+T.parent.x+24)*scale+(float32(offset)*8*scale)),
				int(float32(T.y+T.parent.y)*scale), 11, scale)
		} else {
			Q.Draw_CharScaled(
				int(float32(T.x+T.parent.x+24)*scale+(float32(offset)*8*scale)),
				int(float32(T.y+T.parent.y)*scale), 32, scale)
		}
	}
}

func (T *menuaction_t) draw() {
	Q := T.parent.owner
	scale := Q.scrGetMenuScale()

	if (T.flags & QMF_LEFT_JUSTIFY) != 0 {
		if (T.flags & QMF_GRAYED) != 0 {
			Q.menuDrawStringDark(T.x+T.parent.x+int(LCOLUMN_OFFSET*scale),
				T.y+T.parent.y, T.name)
		} else {
			Q.menuDrawString(T.x+T.parent.x+int(LCOLUMN_OFFSET*scale),
				T.y+T.parent.y, T.name)
		}
	} else {
		if (T.flags & QMF_GRAYED) != 0 {
			Q.menuDrawStringR2LDark(T.x+T.parent.x+int(LCOLUMN_OFFSET*scale),
				T.y+T.parent.y, T.name)
		} else {
			Q.menuDrawStringR2L(T.x+T.parent.x+int(LCOLUMN_OFFSET*scale),
				T.y+T.parent.y, T.name)
		}
	}

	// if (T.ownerdraw) {
	// 	T.ownerdraw(T);
	// }
}

func (T *menuaction_t) doEnter() bool {
	if T.callback != nil {
		T.callback(&T.menucommon_t)
	}
	return true
}

func (T *menulist_t) draw() {
	// char buffer[100];
	Q := T.parent.owner
	scale := Q.scrGetMenuScale()

	if len(T.name) > 0 {
		Q.menuDrawStringR2LDark(T.x+T.parent.x+int(LCOLUMN_OFFSET*scale), T.y+T.parent.y,
			T.name)
	}

	if !strings.Contains(T.itemnames[T.curvalue], "\n") {
		Q.menuDrawString(int(RCOLUMN_OFFSET*scale)+T.x+T.parent.x,
			T.y+T.parent.y, T.itemnames[T.curvalue])
	} else {
		// 	strcpy(buffer, s->itemnames[s->curvalue]);
		// 	*strchr(buffer, '\n') = 0;
		// 	Menu_DrawString(RCOLUMN_OFFSET * scale + s->generic.x +
		// 		   	s->generic.parent->x, s->generic.y +
		// 			s->generic.parent->y, buffer);
		// 	strcpy(buffer, strchr(s->itemnames[s->curvalue], '\n') + 1);
		// 	Menu_DrawString(RCOLUMN_OFFSET * scale + s->generic.x +
		// 			s->generic.parent->x, s->generic.y +
		// 			s->generic.parent->y + 10, buffer);
	}
}

func (T *menulist_t) DoSlide(dir int) {
	T.curvalue += dir

	if T.curvalue < 0 {
		T.curvalue = 0
	} else if T.curvalue >= len(T.itemnames) {
		T.curvalue--
	}

	if T.callback != nil {
		T.callback(&T.menucommon_t)
	}
}

func (T *qClient) menuDrawString(x, y int, str string) {
	scale := T.scrGetMenuScale()

	for i := range str {
		T.Draw_CharScaled(x+int(float32(i*8)*scale), int(float32(y)*scale), int(str[i]), scale)
	}
}

func (T *qClient) menuDrawStringDark(x, y int, str string) {
	scale := T.scrGetMenuScale()

	for i := range str {
		T.Draw_CharScaled(x+int(float32(i*8)*scale), int(float32(y)*scale), int(str[i])+128, scale)
	}
}

func (T *qClient) menuDrawStringR2L(x, y int, str string) {
	scale := T.scrGetMenuScale()

	for i := range str {
		T.Draw_CharScaled(x-int(float32(i*8)*scale), int(float32(y)*scale), int(str[len(str)-i-1]), scale)
	}
}

func (T *qClient) menuDrawStringR2LDark(x, y int, str string) {
	scale := T.scrGetMenuScale()

	for i := range str {
		T.Draw_CharScaled(x-int(float32(i*8)*scale), int(float32(y)*scale), int(str[len(str)-i-1])+128, scale)
	}
}
