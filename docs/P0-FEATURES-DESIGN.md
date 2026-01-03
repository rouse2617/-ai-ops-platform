# AI-Ops P0 功能详细设计文档

## 文档概述

**版本**: 1.0
**日期**: 2026-01-03
**作者**: System Architect
**状态**: 设计阶段

本文档详细描述 AI-Ops 运维平台的两个 P0 级别功能的架构设计和实现方案。

---

## 目录

1. [系统架构概览](#1-系统架构概览)
2. [P0-1: 高危操作确认机制](#2-p0-1-高危操作确认机制)
3. [P0-2: 执行结果标示主机](#3-p0-2-执行结果标示主机)
4. [数据流设计](#4-数据流设计)
5. [核心数据结构](#5-核心数据结构)
6. [接口定义](#6-接口定义)
7. [实现方案](#7-实现方案)
8. [测试用例](#8-测试用例)
9. [风险评估](#9-风险评估)
10. [实施计划](#10-实施计划)

---

## 1. 系统架构概览

### 1.1 当前架构分析

```mermaid
graph TB
    subgraph "前端层 Frontend"
        A[ChatWindow.vue] --> B[MessageList.vue]
        B --> C[EnhancedToolCallCard.vue]
        A --> D[InputBox.vue]
        A --> E[ConfirmationDialog.vue]
    end

    subgraph "状态管理 State"
        F[ChatStore] --> G[Messages]
        F --> H[ToolCalls]
        F --> I[SelectedHosts]
    end

    subgraph "工具层 Utils"
        J[dangerousCommands.ts] --> K[风险检测]
        J --> L[影响评估]
    end

    subgraph "后端层 Backend"
        M[ChatHandler] --> N[ChatService]
        N --> O[Agent]
        O --> P[Tool Registry]
        P --> Q[SSH Executor]
    end

    A --> F
    C --> F
    E --> J
    D --> M
    M --> F
    Q --> R[(主机)]

    style A fill:#e1f5ff
    style E fill:#fff3cd
    style J fill:#f8d7da
    style Q fill:#d4edda
