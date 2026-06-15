#!/bin/bash
show_dir() {
    local dir="$1"
    local prefix="$2"
    local files=($(ls -1 "$dir" 2>/dev/null | grep -v "^$\|^\." | head -3))
    local all_files=($(ls -1 "$dir" 2>/dev/null | grep -v "^$\|\."))
    
    for file in "${files[@]}"; do
        if [ -d "$dir/$file" ]; then
            echo "${prefix}├── 📁 $file/"
            show_dir "$dir/$file" "${prefix}│   "
        else
            echo "${prefix}├── 📄 $file"
        fi
    done
    
    if [ ${#all_files[@]} -gt 3 ]; then
        echo "${prefix}└── ... and $((${#all_files[@]} - 3)) more items"
    fi
}

echo "📁 $(basename $(pwd))/"
show_dir "." "    "
