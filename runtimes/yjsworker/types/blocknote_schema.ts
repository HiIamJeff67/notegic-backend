import {
  type Block,
  BlockNoteEditor,
  BlockNoteSchema,
  createBlockSpec,
  createInlineContentSpec,
  defaultBlockSpecs,
  defaultInlineContentSpecs,
} from "@blocknote/core";

const { divider: _, ...notegicBlockSpecs } = defaultBlockSpecs;

const createPlainBlockSpec = <T extends "mathBlock" | "diagram">(type: T) =>
  createBlockSpec(
    {
      type,
      propSchema: {},
      content: "plain",
    },
    {
      render: () => {
        const dom = document.createElement("div");
        return { dom, contentDOM: dom };
      },
    }
  )();

const calendarBlockSpec = createBlockSpec(
  {
    type: "calendar",
    propSchema: {
      calendarId: { default: "" },
      anchorDate: { default: "" },
      timezone: { default: "UTC" },
      view: { default: "month", values: ["month"] as const },
    },
    content: "none",
  },
  {
    render: () => ({ dom: document.createElement("div") }),
  }
)();

const mathInlineContentSpec = createInlineContentSpec(
  {
    type: "math",
    propSchema: {},
    content: "plain",
  },
  {
    render: () => {
      const dom = document.createElement("span");
      const contentDOM = document.createElement("span");
      dom.append(contentDOM);
      return { dom, contentDOM };
    },
  }
);

export const notegicBlockNoteSchema = BlockNoteSchema.create({
  blockSpecs: {
    ...notegicBlockSpecs,
    mathBlock: createPlainBlockSpec("mathBlock"),
    diagram: createPlainBlockSpec("diagram"),
    calendar: calendarBlockSpec,
  },
  inlineContentSpecs: {
    ...defaultInlineContentSpecs,
    math: mathInlineContentSpec,
  },
});

export const notegicBlockNoteEditor = BlockNoteEditor.create({
  schema: notegicBlockNoteSchema,
});

export type NotegicBlock = Block<
  typeof notegicBlockNoteSchema.blockSchema,
  typeof notegicBlockNoteSchema.inlineContentSchema,
  typeof notegicBlockNoteSchema.styleSchema
>;
