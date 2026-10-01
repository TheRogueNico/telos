# Design

## Overview

The application organizes work into **Groups**, **Tasks**, and **Key Tasks**.

The following describes the structure of the data.
It is **not** the database implementation; see the [Database](#database) section for that.

### Groups

A group is a collection of related tasks.

```
Group
 ├── ID
 └── Name*
```

### Tasks

A task represents a specific objective to be accomplished.

```
Task
 ├── ID
 ├── Title*
 ├── Intent*
 ├── Key Tasks
 ├── End State*
 ├── Status
 └── Creation Date
```

* **Intent** - What the task must achieve.
* **Key Tasks** - What must be accomplished.
* **End State** - What success looks like when the task is done.

A task belongs to one group and may contain multiple key tasks.

### Key Tasks

A key task describes a specific action or objective required to accomplish its parent task.

```text
Key Task
 ├── ID
 ├── Description*
 └── Status
```

## Defaults and Required Fields

Fields marked with `*` must be specified by the user when the item is created.

ID fields are automatically generated.

For status fields, the application assigns a default value to them at creation:

| Entity   | Field  | Allowed Values                              | Default         |
| -------- | ------ | ------------------------------------------- | --------------- |
| Task     | Status | `active`, `paused`, `canceled`, `completed` | `active`        |
| Key Task | Status | `completed`, `not completed`                | `not completed` |

## Database

The structure above is represented in the database as follows.

### Group

| Column | Constraint  |
| ------ | ----------- |
| `ID`   | Primary Key |
| `Name` | Not Null    |

### Task

| Column         | Constraint               |
| -------------- | ------------------------ |
| `ID`           | Primary Key              |
| `GroupID`      | Foreign Key              |
| `Title`        | Not Null                 |
| `Intent`       | Not Null                 |
| `EndState`     | Not Null                 |
| `Status`       | Not Null                 |
| `CreationDate` | Not Null                 |

### KeyTask

| Column        | Constraint              |
| ------------- | ----------------------- |
| `ID`          | Primary Key             |
| `TaskID`      | Foreign Key             |
| `Description` | Not Null                |
| `Status`      | Not Null                |

### Relationships

```
Group --[1:N]-> Task --[1:N]-> KeyTask
```

* A **Group** can contain multiple **Tasks**.
* A **Task** belongs to one **Group**.
* A **Task** can contain multiple **Key Tasks**.
* A **Key Task** belongs to one **Task**.
