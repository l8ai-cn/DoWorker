import { useEffect, useRef, useState } from "react";
import { CheckIcon, ChevronDownIcon, PlusIcon, SearchIcon, TagIcon } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { useProjects } from "@/hooks/useConversations";

export function LandingProjectPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (project: string) => void;
}) {
  const { data: projects = [] } = useProjects();
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [creatingNew, setCreatingNew] = useState(false);
  const [newName, setNewName] = useState("");
  const newRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (creatingNew) newRef.current?.focus();
  }, [creatingNew]);

  const filtered = search
    ? projects.filter((name) => name.toLowerCase().includes(search.toLowerCase()))
    : projects;
  const itemClass =
    "flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-xs hover:bg-accent hover:text-accent-foreground";

  function pick(project: string) {
    onChange(project);
    setOpen(false);
    setSearch("");
    setCreatingNew(false);
    setNewName("");
  }

  function commitNew() {
    const name = newName.trim();
    if (name) pick(name);
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button
          type="button"
          className="flex h-6 items-center gap-1 rounded-full px-2.5 text-13 font-normal text-muted-foreground transition-colors hover:text-foreground"
          data-testid="new-chat-landing-project-chip"
        >
          <TagIcon className="size-4 shrink-0" />
          <span className={`hidden max-w-32 truncate sm:block ${value ? "text-foreground" : ""}`}>
            {value || "No project"}
          </span>
          <ChevronDownIcon className="size-3.5 shrink-0 opacity-60" />
        </button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-56 p-1" onCloseAutoFocus={(e) => e.preventDefault()}>
        <div className="flex items-center gap-2 border-b px-2 py-1.5">
          <SearchIcon className="size-3.5 shrink-0 text-muted-foreground" />
          <input
            className="w-full bg-transparent text-xs outline-none placeholder:text-muted-foreground"
            placeholder="Search projects"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="max-h-48 overflow-y-auto">
          <ProjectOption label="No project" selected={value === ""} onClick={() => pick("")} />
          {filtered.map((name) => (
            <ProjectOption
              key={name}
              label={name}
              selected={value === name}
              onClick={() => pick(name)}
            />
          ))}
          {filtered.length === 0 && !creatingNew && (
            <p className="px-2 py-1.5 text-xs text-muted-foreground">No projects yet.</p>
          )}
        </div>
        <div className="border-t pt-1">
          {creatingNew ? (
            <input
              ref={newRef}
              className="w-full bg-transparent px-2 py-1 text-xs outline-none"
              placeholder="Project name…"
              value={newName}
              onChange={(event) => setNewName(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  commitNew();
                }
                if (event.key === "Escape") {
                  setCreatingNew(false);
                  setNewName("");
                }
              }}
            />
          ) : (
            <button type="button" className={itemClass} onClick={() => setCreatingNew(true)}>
              <PlusIcon className="size-3.5 shrink-0" />
              New project…
            </button>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}

function ProjectOption({
  label,
  selected,
  onClick,
}: {
  label: string;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button type="button" className="flex w-full items-center gap-1.5 rounded-md px-2 py-1.5 text-left text-xs hover:bg-accent hover:text-accent-foreground" onClick={onClick}>
      <span className="flex-1 truncate">{label}</span>
      {selected && <CheckIcon className="size-3.5 shrink-0 text-primary" />}
    </button>
  );
}
