#!/usr/bin/env python3
import os
import sys
import argparse
import re

def get_module_name():
    """Attempt to auto-detect Go module name from go.mod"""
    paths_to_check = ["go.mod", "backend/go.mod", "../go.mod"]
    for path in paths_to_check:
        if os.path.exists(path):
            try:
                with open(path, "r") as f:
                    content = f.read()
                    match = re.search(r"module\s+(.+)", content)
                    if match:
                        return match.group(1).strip()
            except Exception:
                continue
    return "your-project"

def load_template(name, default_content):
    """Load template from file if exists, otherwise use default"""
    template_path = os.path.join(os.path.dirname(__file__), "..", "templates", f"{name}.go.tmpl")
    if os.path.exists(template_path):
        try:
            with open(template_path, 'r') as f:
                return f.read()
        except Exception:
            pass
    return default_content

# Fallback Internal Templates (for portability)
DEFAULT_HANDLER = """package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "{module}/internal/dto"
    "{module}/internal/service"
    "{module}/pkg/response"
    "{module}/pkg/apperror"
)

type {domain_title}Handler struct {{
    svc service.{domain_title}Service
}}

func New{domain_title}Handler(svc service.{domain_title}Service) *{domain_title}Handler {{
    return &{domain_title}Handler{{svc: svc}}
}}

func (h *{domain_title}Handler) GetByID(c *gin.Context) {{
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {{
        response.Fail(c, http.StatusBadRequest, apperror.BadRequest("invalid id"))
        return
    }}

    res, appErr := h.svc.GetByID(c.Request.Context(), uint(id))
    if appErr != nil {{
        response.Fail(c, appErr.HTTPStatus, appErr)
        return
    }}

    response.OK(c, res)
}}
"""

DEFAULT_SERVICE_IMPL = """package service

import (
    "context"
    "{module}/internal/dto"
    "{module}/internal/repository"
    "{module}/pkg/apperror"
)

type {domain_lower}ServiceImpl struct {{
    repo repository.{domain_title}Repository
}}

func New{domain_title}Service(repo repository.{domain_title}Repository) {domain_title}Service {{
    return &{domain_lower}ServiceImpl{{repo: repo}}
}}

func (s *{domain_lower}ServiceImpl) GetByID(ctx context.Context, id uint) (*dto.{domain_title}Response, *apperror.AppError) {{
    entity, err := s.repo.FindByID(ctx, id)
    if err != nil {{
        return nil, apperror.NotFound("{domain_lower} not found")
    }}
    return dto.To{domain_title}Response(entity), nil
}}
"""

SERVICE_INTERFACE_SNIPPET = """
type {domain_title}Service interface {{
    GetByID(ctx context.Context, id uint) (*dto.{domain_title}Response, *apperror.AppError)
}}
"""

DEFAULT_REPO_IMPL = """package repository

import (
    "context"
    "{module}/internal/model"
    "gorm.io/gorm"
)

type {domain_lower}RepositoryImpl struct {{
    db *gorm.DB
}}

func New{domain_title}Repository(db *gorm.DB) {domain_title}Repository {{
    return &{domain_lower}RepositoryImpl{{db: db}}
}}

func (r *{domain_lower}RepositoryImpl) FindByID(ctx context.Context, id uint) (*model.{domain_title}, error) {{
    var entity model.{domain_title}
    if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {{
        return nil, err
    }}
    return &entity, nil
}}
"""

REPO_INTERFACE_SNIPPET = """
type {domain_title}Repository interface {{
    FindByID(ctx context.Context, id uint) (*model.{domain_title}, error)
}}
"""

DEFAULT_MODEL = """package model

import "time"

type {domain_title} struct {{
    BaseModel
    Name        string `gorm:"type:varchar(255);not null"`
    Description string `gorm:"type:text"`
}}

func ({domain_title}) TableName() string {{
    return "{domain_lower}s"
}}
"""

DEFAULT_DTO = """package dto

import (
    "time"
    "{module}/internal/model"
)

type {domain_title}Response struct {{
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}}

func To{domain_title}Response(m *model.{domain_title}) *{domain_title}Response {{
    return &{domain_title}Response{{
        ID:          m.ID,
        Name:        m.Name,
        Description: m.Description,
        CreatedAt:   m.CreatedAt,
    }}
}}
"""

def update_interfaces(file_path, package_name, interfaces_content, imports=[]):
    """Safely update or create interfaces.go"""
    if not os.path.exists(file_path):
        os.makedirs(os.path.dirname(file_path), exist_ok=True)
        with open(file_path, 'w') as f:
            f.write(f"package {package_name}\n\n")
            if imports:
                f.write("import (\n")
                for imp in imports:
                    f.write(f'    "{imp}"\n')
                f.write(")\n\n")
            f.write(interfaces_content)
    else:
        with open(file_path, 'r') as f:
            content = f.read()
            if interfaces_content.strip() in content:
                print(f"  ℹ️  Interfaces already present in {file_path}")
                return
        with open(file_path, 'a') as f:
            f.write(interfaces_content)

def scaffold(domain, base_dir, module):
    domain_title = domain.capitalize()
    domain_lower = domain.lower()
    
    # Path settings
    h_dir = f"{base_dir}/handler"
    s_dir = f"{base_dir}/service"
    r_dir = f"{base_dir}/repository"
    m_dir = f"{base_dir}/model"
    d_dir = f"{base_dir}/dto"

    ctx = {
        "module": module,
        "domain_title": domain_title,
        "domain_lower": domain_lower
    }
    
    files = [
        (f"{h_dir}/{domain_lower}_handler.go", load_template("handler", DEFAULT_HANDLER)),
        (f"{s_dir}/{domain_lower}_service.go", load_template("service_impl", DEFAULT_SERVICE_IMPL)),
        (f"{r_dir}/{domain_lower}_repository.go", load_template("repository_impl", DEFAULT_REPO_IMPL)),
        (f"{m_dir}/{domain_lower}.go", load_template("model", DEFAULT_MODEL)),
        (f"{d_dir}/{domain_lower}_dto.go", load_template("dto", DEFAULT_DTO)),
    ]
    
    print(f"🏗️  Scaffolding {domain_title} domain (Module: {module})...")
    
    for path, template in files:
        os.makedirs(os.path.dirname(path), exist_ok=True)
        content = template.format(**ctx)
        with open(path, 'w') as f:
            f.write(content)
        print(f"  ✅ Created: {path}")

    # Update interfaces
    svc_info = ["context", f"{module}/{d_dir}", f"{module}/pkg/apperror"]
    update_interfaces(f"{s_dir}/interfaces.go", "service", SERVICE_INTERFACE_SNIPPET.format(**ctx), svc_info)
    
    repo_info = ["context", f"{module}/{m_dir}"]
    update_interfaces(f"{r_dir}/interfaces.go", "repository", REPO_INTERFACE_SNIPPET.format(**ctx), repo_info)
    print(f"  ✅ Synchronized interfaces for {domain_title}")

def main():
    parser = argparse.ArgumentParser(description='Go Backend Scaffolder (Wave 5: Modular)')
    parser.add_argument('--domain', required=True, help='Domain name (e.g., user)')
    parser.add_argument('--base', default='internal', help='Base directory (default: internal)')
    parser.add_argument('--module', help='Go module name (auto-detected if not provided)')
    args = parser.parse_args()
    
    module = args.module or get_module_name()
    scaffold(args.domain, args.base.strip("/"), module)
    print("\n🚀 Scaffolding complete! Modular templates used if available.")

if __name__ == "__main__":
    main()
