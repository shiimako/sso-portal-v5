package models

type StudyProgram struct {
	ID   int    `db:"id"`
	Study_Program_name string `db:"study_program_name"`
	MajorID int    `db:"major_id"`
}