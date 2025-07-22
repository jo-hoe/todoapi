package todoclient

import (
	"testing"
	"time"
)

func TestToDoTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    ToDoTask
		wantErr bool
	}{
		{
			name: "valid task",
			task: ToDoTask{
				ID:          "1",
				Name:        "Test task",
				Description: "Test description",
				DueDate:     time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "empty name",
			task: ToDoTask{
				ID:          "1",
				Name:        "",
				Description: "Test description",
			},
			wantErr: true,
		},
		{
			name: "name too long",
			task: ToDoTask{
				ID:   "1",
				Name: string(make([]byte, 501)),
			},
			wantErr: true,
		},
		{
			name: "description too long",
			task: ToDoTask{
				ID:          "1",
				Name:        "Test task",
				Description: string(make([]byte, 2001)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToDoTask.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToDoParent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		parent  ToDoParent
		wantErr bool
	}{
		{
			name: "valid parent",
			parent: ToDoParent{
				ID:   "1",
				Name: "Test project",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			parent: ToDoParent{
				ID:   "1",
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "name too long",
			parent: ToDoParent{
				ID:   "1",
				Name: string(make([]byte, 201)),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parent.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToDoParent.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
