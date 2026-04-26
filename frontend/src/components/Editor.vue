<template>
  <div class="editor-wrapper border border-slate-200 rounded-xl overflow-hidden">
    <!-- Toolbar -->
    <div v-if="editor" class="editor-toolbar bg-slate-50 border-b border-slate-200 px-3 py-2 flex flex-wrap items-center gap-1">
      <!-- Text Style -->
      <div class="flex items-center gap-0.5">
        <ToolBtn @click="editor.chain().focus().toggleBold().run()" :active="editor.isActive('bold')" title="加粗">
          <Bold class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleItalic().run()" :active="editor.isActive('italic')" title="斜体">
          <Italic class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleStrike().run()" :active="editor.isActive('strike')" title="删除线">
          <Strikethrough class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleHighlight().run()" :active="editor.isActive('highlight')" title="高亮">
          <Highlighter class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleCode().run()" :active="editor.isActive('code')" title="行内代码">
          <Code class="w-3.5 h-3.5" />
        </ToolBtn>
      </div>

      <div class="w-px h-5 bg-slate-200 mx-1"></div>

      <!-- Headings -->
      <div class="flex items-center gap-0.5">
        <ToolBtn @click="editor.chain().focus().toggleHeading({ level: 1 }).run()" :active="editor.isActive('heading', { level: 1 })" title="标题 1">
          <Heading1 class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleHeading({ level: 2 }).run()" :active="editor.isActive('heading', { level: 2 })" title="标题 2">
          <Heading2 class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleHeading({ level: 3 }).run()" :active="editor.isActive('heading', { level: 3 })" title="标题 3">
          <Heading3 class="w-3.5 h-3.5" />
        </ToolBtn>
      </div>

      <div class="w-px h-5 bg-slate-200 mx-1"></div>

      <!-- Lists -->
      <div class="flex items-center gap-0.5">
        <ToolBtn @click="editor.chain().focus().toggleBulletList().run()" :active="editor.isActive('bulletList')" title="无序列表">
          <List class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleOrderedList().run()" :active="editor.isActive('orderedList')" title="有序列表">
          <ListOrdered class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleTaskList().run()" :active="editor.isActive('taskList')" title="任务列表">
          <ListChecks class="w-3.5 h-3.5" />
        </ToolBtn>
      </div>

      <div class="w-px h-5 bg-slate-200 mx-1"></div>

      <!-- Block -->
      <div class="flex items-center gap-0.5">
        <ToolBtn @click="editor.chain().focus().toggleBlockquote().run()" :active="editor.isActive('blockquote')" title="引用">
          <Quote class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().toggleCodeBlock().run()" :active="editor.isActive('codeBlock')" title="代码块">
          <Code2 class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().setHorizontalRule().run()" title="分割线">
          <Minus class="w-3.5 h-3.5" />
        </ToolBtn>
      </div>

      <div class="w-px h-5 bg-slate-200 mx-1"></div>

      <!-- Table -->
      <div class="flex items-center gap-0.5">
        <ToolBtn @click="insertTable" title="插入表格">
          <Table class="w-3.5 h-3.5" />
        </ToolBtn>
      </div>

      <!-- Link -->
      <ToolBtn @click="setLink" :active="editor.isActive('link')" title="链接">
        <Link class="w-3.5 h-3.5" />
      </ToolBtn>

      <!-- Image -->
      <ToolBtn @click="triggerImageUpload" title="插入图片">
        <ImageIcon class="w-3.5 h-3.5" />
      </ToolBtn>
      <input ref="imageInput" type="file" accept="image/*" class="hidden" @change="handleImageUpload" />

      <div class="w-px h-5 bg-slate-200 mx-1"></div>

      <!-- History -->
      <div class="flex items-center gap-0.5">
        <ToolBtn @click="editor.chain().focus().undo().run()" :disabled="!editor.can().undo()" title="撤销">
          <Undo2 class="w-3.5 h-3.5" />
        </ToolBtn>
        <ToolBtn @click="editor.chain().focus().redo().run()" :disabled="!editor.can().redo()" title="重做">
          <Redo2 class="w-3.5 h-3.5" />
        </ToolBtn>
      </div>

      <!-- Character count -->
      <div class="ml-auto text-xs text-slate-400">
        {{ editor.storage.characterCount?.characters() || 0 }} 字符
      </div>
    </div>

    <!-- Editor content -->
    <editor-content :editor="editor" class="prose prose-slate max-w-none p-4 min-h-64 focus:outline-none" />
  </div>
</template>

<script setup>
import { ref, watch, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
import Placeholder from '@tiptap/extension-placeholder'
import Table from '@tiptap/extension-table'
import TableRow from '@tiptap/extension-table-row'
import TableCell from '@tiptap/extension-table-cell'
import TableHeader from '@tiptap/extension-table-header'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'
import Highlight from '@tiptap/extension-highlight'
import TextStyle from '@tiptap/extension-text-style'
import CharacterCount from '@tiptap/extension-character-count'
import { fileApi } from '@/api/index.js'

import {
  Bold, Italic, Strikethrough, Code, Code2, Heading1, Heading2, Heading3,
  List, ListOrdered, ListChecks, Quote, Minus, Table as TableIcon,
  Link as LinkIcon, Image as ImageIcon, Undo2, Redo2, Highlighter
} from 'lucide-vue-next'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '开始编写...' }
})
const emit = defineEmits(['update:modelValue'])
const imageInput = ref(null)

// Rename icon aliases
const Table2 = TableIcon
const Link2 = LinkIcon

const editor = useEditor({
  content: props.modelValue,
  extensions: [
    StarterKit,
    Image.configure({ allowBase64: true }),
    Link.configure({ openOnClick: false }),
    Placeholder.configure({ placeholder: props.placeholder }),
    Table.configure({ resizable: true }),
    TableRow,
    TableCell,
    TableHeader,
    TaskList,
    TaskItem.configure({ nested: true }),
    Highlight,
    TextStyle,
    CharacterCount
  ],
  onUpdate({ editor }) {
    emit('update:modelValue', editor.getHTML())
  }
})

watch(() => props.modelValue, (val) => {
  if (editor.value && editor.value.getHTML() !== val) {
    editor.value.commands.setContent(val, false)
  }
})

function insertTable() {
  editor.value?.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
}

function setLink() {
  const prev = editor.value?.getAttributes('link').href
  const url = window.prompt('输入链接地址', prev || 'https://')
  if (url === null) return
  if (url === '') {
    editor.value?.chain().focus().extendMarkRange('link').unsetLink().run()
    return
  }
  editor.value?.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
}

function triggerImageUpload() {
  imageInput.value?.click()
}

async function handleImageUpload(e) {
  const file = e.target.files[0]
  if (!file) return
  try {
    const fd = new FormData()
    fd.append('file', file)
    const res = await fileApi.upload(fd)
    const url = res.data.data?.url
    if (url) {
      editor.value?.chain().focus().setImage({ src: url }).run()
    }
  } catch (err) {
    alert('图片上传失败')
  }
  e.target.value = ''
}

onBeforeUnmount(() => {
  editor.value?.destroy()
})

// Subcomponent for toolbar button
</script>

<script>
import { defineComponent, h } from 'vue'
const ToolBtn = defineComponent({
  props: { active: Boolean, disabled: Boolean, title: String },
  emits: ['click'],
  setup(props, { slots, emit }) {
    return () => h('button', {
      type: 'button',
      title: props.title,
      disabled: props.disabled,
      onClick: () => emit('click'),
      class: [
        'p-1.5 rounded transition-all',
        props.active ? 'bg-indigo-100 text-indigo-700' : 'text-slate-500 hover:bg-slate-200 hover:text-slate-700',
        props.disabled ? 'opacity-30 cursor-not-allowed' : ''
      ].join(' ')
    }, slots.default?.())
  }
})
export { ToolBtn }
</script>
